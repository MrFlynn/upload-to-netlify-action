const core = require("@actions/core");
const crypto = require("crypto");
const fs = require("fs");
const netlify = require("netlify");

async function getSiteIDFromName(client) {
  const name = core.getInput("site-name");
  const site = await client.listSites({ name });

  return site["site_id"];
}

async function pollDeploy(client, siteID) {
  const deploys = await client.listSiteDeploys({ site_id: siteID });
  const deployID = deploys[0]["id"];

  while (true) {
    site = await client.getSiteDeploy({ site_id: siteID, deploy_id: deployID });
    if (site["locked"] === true || site["id"] === null) {
      return deployID;
    }

    await new Promise((r) => setTimeout(r, 5000));
  }
}

try {
  // Get token. Ensure it's masked from any logs.
  const token = core.getInput("netlify-token");
  core.setSecret(token);

  const client = new netlify(token);

  // Get the site ID from the name given to the action. Wait for deploy to finish.
  const siteID = await getSiteIDFromName(client);
  const deployID = await pollDeploy(client, siteID);

  // Read file contents and generate SHA1 hash.
  const sourceFile = core.getInput("source-file");
  var stream = fs.createReadStream(sourceFile);

  var shasum = crypto.createHash("sha1");
  var digest = undefined;
  var size = 0;

  // Read filestream and hash all chunks.
  stream
    .on("data", (chunk) => {
      shasum.update(chunk);
      size += chunk.size;
    })
    .on("end", (_) => {
      digest = shasum.digest("base64");
    });

  // Update the deploy with new file information.
  const destination = core.getInput("destination-path");
  await client.updateSiteDeploy({
    site_id: siteID,
    deploy_id: deployID,
    body: {
      files: {
        destination: digest,
      },
    },
  });

  core.info(`Uploading file ${sourceFile} (${size} bytes)...`);

  // Upload file.
  await client.uploadDeployFile({
    site_id: siteID,
    deploy_id: deployID,
    size,
    body: fs.createReadStream(sourceFile),
  });

  core.info("File successfully uploaded to Netlify!");
} catch (e) {
  core.setFailed(error.message);
}
