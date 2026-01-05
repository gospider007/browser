const fingerprint_generator = require("fingerprint-generator");
const generator = new fingerprint_generator.FingerprintGenerator();
c=generator.getFingerprint({
    browsers: [
        { name: "edge"},
    ],
    operatingSystems: ["android"],
        strict: true,
});
console.log(c)


