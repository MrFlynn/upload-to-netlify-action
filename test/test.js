const os = require("os");

const assert = require("chai").assert;
const expect = require("chai").expect;
const sinon = require("sinon");
const proxyquire = require("proxyquire");

const fsify = require("fsify")({
  cwd: os.tmpdir(),
  persistent: false,
  force: true,
});

const utils = proxyquire("../index.js", {
  core: {
    getInput: function (arg) {
      return "example-site";
    },
    "@noCallThru": true,
  },
});

const client = {
  listSites: function () {},
  listSiteDeploys: function () {},
  getSiteDeploy: function () {
    return {
      state: "ready",
    };
  },
  listSiteFiles: function () {},
};

var sandbox = sinon.createSandbox();

describe("Utilities tests", function () {
  this.beforeEach(function () {
    sandbox = sinon.createSandbox();
  });

  this.afterEach(function () {
    sandbox.restore();
  });

  describe("getSiteIDFromName", function () {
    it("should return the correct site ID", async function () {
      const listSitesStub = sandbox.stub(client, "listSites").returns([
        {
          site_id: "example-id",
        },
      ]);

      const site_id = await utils.getSiteIDFromName(client);

      assert.equal(site_id, "example-id");
      expect(listSitesStub.calledOnce).to.be.true;
    });
  });

  describe("pollDeploy", function () {
    const listSiteDeploysStub = sandbox
      .stub(client, "listSiteDeploys")
      .returns([
        {
          id: "example-deploy-id",
        },
      ]);

    it("should return immediately", async function () {
      const getSiteDeploySpy = sandbox.spy(client, "getSiteDeploy");

      await utils.pollDeploy(client, "example-site-id");

      expect(listSiteDeploysStub.calledOnce).to.be.true;
      expect(getSiteDeploySpy.calledOnce).to.be.true;
      expect(getSiteDeploySpy.args[0]).to.deep.equal([
        { site_id: "example-site-id", deploy_id: "example-deploy-id" },
      ]);
    });

    it("should throw error if deploy state has failed", async function () {
      const getSiteDeployStub = sandbox.stub(client, "getSiteDeploy").returns({
        state: "error",
      });

      var err;
      try {
        await utils.pollDeploy(client, "sample-site-id");
      } catch (e) {
        err = e;
      }

      assert.instanceOf(err, Error);
      expect(err.message).to.equal(
        "Existing build failed. Terminating upload..."
      );
    });
  });

  describe("getHash", function () {
    const structure = [
      {
        type: fsify.FILE,
        name: "test.txt",
        contents: "hello, world!",
      },
    ];

    it("should return file hash", function () {
      return fsify(structure).then(
        (struct) =>
          new Promise(async (resolve) => {
            const hash = await utils.getHash(struct[0].name);
            resolve(
              assert.equal(hash, "1f09d30c707d53f3d16c530dd73d70a6ce7596a9")
            );
          })
      );
    });
  });

  describe("cleanDestinationPath", function () {
    it("should return the original string", function () {
      assert.equal(utils.cleanDestinationPath("test"), "test");
    });

    it("should remove the leading /", function () {
      assert.equal(utils.cleanDestinationPath("/test"), "test");
    });

    it("should throw an error with '#?' in path", function () {
      expect(() => utils.cleanDestinationPath("hello#?world").to.throw(Error));
    });
  });

  describe("getFileList", function () {
    it("should replace proper sha in the file list", async function () {
      let expectation = {
        "/index.html": "abc",
        "/asset.pdf": "ghi",
      };

      const listSiteFilesStub = sandbox.stub(client, "listSiteFiles").returns([
        {
          id: "/index.html",
          sha: "abc",
        },
        {
          id: "/asset.pdf",
          sha: "def",
        },
      ]);

      const files = await utils.getFileList(
        client,
        "/asset.pdf",
        "ghi",
        "example"
      );

      assert.deepEqual(files, expectation);
      expect(listSiteFilesStub.calledOnce).to.be.true;
    });
  });
});
