const crypto = require("node:crypto");

/**
 * Vite 7 relies on the new crypto.hash API which is only available in Node 20.19+.
 * Provide a backwards compatible shim so we can build with slightly older LTS releases.
 */
if (typeof crypto.hash !== "function") {
  crypto.hash = (algorithm, data, encoding = "hex") => {
    const hash = crypto.createHash(algorithm);
    hash.update(data);

    if (encoding === "buffer") {
      return hash.digest();
    }

    return hash.digest(encoding);
  };
}
