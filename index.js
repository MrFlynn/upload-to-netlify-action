const core = require("@actions/core");
const crypto = require("crypto");
const fs = require("fs");
const netlify = require("netlify");

// Get token and create new Netlify instance.
token = core.getInput("netlify-token");
core.setSecret(token);

const client = new netlify(token);

const getSiteIDFromName = async () => {
  const name = core.getInput("site-name");
  const site = await client.listSites({ name });

  return site[0]["site_id"];
};

const pollDeploy = async (site_id) => {
  const deploys = await client.listSiteDeploys({ site_id });
  const deploy_id = deploys[0]["id"];

  while (true) {
    deploy = await client.getSiteDeploy({ site_id, deploy_id });

    if (deploy["state"] === "error") {
      throw "Existing build failed. Terminating upload...";
    } else if (deploy["state"] === "ready") {
      break;
    }

    await new Promise((r) => setTimeout(r, 5000));
  }
};

const getHash = (path) =>
  new Promise((resolve) => {
    const hash = crypto.createHash("sha1");
    fs.createReadStream(path)
      .on("data", (chunk) => {
        hash.update(chunk);
      })
      .on("end", () => {
        resolve(hash.digest("hex"));
      });
  });

const cleanDestinationPath = (path) => {
  if ((chars = path.match(/[#?]/g)) !== null) {
    throw `Destination path contains illegal characters: ${chars.join(", ")}`;
  }

  if (path.startsWith("/")) {
    return path.slice(1);
  } else {
    return path;
  }
};

(async () => {
  try {
    const source_file = core.getInput("source-file");
    core.info(`Uploading file ${source_file} to Netlify...`);

    // Get the site ID from the name given to the action. Wait for current deploy to finish.
    const site_id = await getSiteIDFromName();
    await pollDeploy(site_id);

    // Get hash of file.
    const digest = await getHash(source_file);

    // Update the deploy with new file information.
    const destination = cleanDestinationPath(core.getInput("destination-path"));
    const deploy = await client.createSiteDeploy({
      site_id,
      body: {
        files: {
          [destination]: digest,
        },
      },
    });

    // Upload file.
    await client.uploadDeployFile({
      deploy_id: deploy["id"],
      path: destination,
      body: fs.createReadStream(source_file),
    });

    core.info("File successfully uploaded to Netlify!");
  } catch (e) {
    core.setFailed(e.message);
  }
})();
