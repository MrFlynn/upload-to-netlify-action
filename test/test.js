const utils = require("../index.js");

const os = require("os");

const assert = require("chai").assert;
const expect = require("chai").expect;
const fsify = require("fsify")({
  cwd: os.tmpdir(),
  persistent: false,
  force: true,
});

describe("Utilities tests", function () {
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
});
