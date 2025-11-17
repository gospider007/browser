import { HeaderGenerator } from 'header-generator';
var g = new HeaderGenerator()
var data = g.getHeaders({
    locales: ["en-US"],
    browsers: ["chrome"],
    devices: ["desktop"],
    operatingSystems: ["macos"],
    mockWebRTC: true,
});
// console.log(data);


