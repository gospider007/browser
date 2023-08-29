const fs = require('fs');
const { FingerprintInjector } = require('fingerprint-injector');
const { FingerprintGenerator } = require('fingerprint-generator');
const injector = new FingerprintInjector();
const generator = new FingerprintGenerator();
const fingerprint = generator.getFingerprint({
    'devices': ['desktop'],
    'operatingSystems': ['windows'],
});
injectable_code=injector.getInjectableScript(fingerprint)
fs.writeFile('stealth2.js', injectable_code, (err) => {
    if (err) {
        console.error('保存文件时出错:', err);
        return;
    }
    console.log('文件已保存成功！');
});


