const { FingerprintInjector } = require('fingerprint-injector');
const fingerprint_generator_1 = require("fingerprint-generator");
const injector = new FingerprintInjector();
const beautify = require("js-beautify").js;

function prettierJs(code){
    code = beautify(code, { indent_size: 2 });
    return code
}





function inject2() {
    const changeCanvasMain=function() {
        function random(list) {
          let min = 0;
          let max = list.length
          return list[Math.floor(Math.random() * (max - min)) + min];
        }
        let rsalt = random([...Array(7).keys()].map(a => a - 3))
        let gsalt = random([...Array(7).keys()].map(a => a - 3))
        let bsalt = random([...Array(7).keys()].map(a => a - 3))
        let asalt = random([...Array(7).keys()].map(a => a - 3))
        let ctxArr = [];
        overridePropertyWithProxy(CanvasRenderingContext2D.prototype, 'getImageData', {
            apply: function (target, ctx, args) {
                const imageData=cache.Reflect.apply(target, ctx, args);
                if (ctxArr.indexOf(ctx)!=-1){
                    return imageData
                }
                let width=imageData.width
                let height=imageData.height
                for (let i = 0; i < height;i++) {
                    for (let j = 0; j < width;j++) {
                        const n = i * (width * 4) + (j * 4);
                        imageData.data[n + 0] = imageData.data[n + 0] + rsalt;
                        imageData.data[n + 1] = imageData.data[n + 1] + gsalt;
                        imageData.data[n + 2] = imageData.data[n + 2] + bsalt;
                        imageData.data[n + 3] = imageData.data[n + 3] + asalt;
                    }
                }
                ctx.putImageData(imageData, 0, 0);
                ctxArr.push(ctx)
                return imageData
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        });
        overridePropertyWithProxy(HTMLCanvasElement.prototype, 'toBlob', {
            apply: function (target, ctx, args) {
                var canvase=ctx.getContext("2d")
                if (canvase){
                    canvase.getImageData(0, 0,  canvase.canvas.width, canvase.canvas.height)
                }
                return cache.Reflect.apply(target, ctx, args)
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        });
        overridePropertyWithProxy(HTMLCanvasElement.prototype, 'toDataURL', {
            apply: function (target, ctx, args) {
                var canvase=ctx.getContext("2d")
                if (canvase){
                    canvase.getImageData(0, 0,  canvase.canvas.width, canvase.canvas.height)
                }
                return cache.Reflect.apply(target, ctx, args)
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        });
    };
    const changeWebGlMain=function () {
        const bufferData = {
            apply: function (target, ctx, args) {
                var index = Math.floor(Math.random() * arguments[1].length);
                var noise = arguments[1][index] !== undefined ? 0.1 * Math.random() * arguments[1][index] : 0;
                arguments[1][index] = arguments[1][index] + noise;
                return cache.Reflect.apply(target, ctx, args);
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        };
        overridePropertyWithProxy(WebGLRenderingContext.prototype, 'bufferData',bufferData);
        overridePropertyWithProxy(WebGL2RenderingContext.prototype, 'bufferData',bufferData);
    };
    const changeAudio = function() {
        var BUFFER=null
        overridePropertyWithProxy(AudioBuffer.prototype, 'getChannelData',{
            apply: function (target, ctx, args) {
                const results_1 = cache.Reflect.apply(target, ctx, args);
                if (BUFFER !== results_1) {
                  BUFFER = results_1;
                  for (var i = 0; i < results_1.length; i += 100) {
                    let index = Math.floor(Math.random() * i);
                    results_1[index] = results_1[index] + Math.random() * 0.0000001;
                  }
                }
                return results_1;
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        });
        var BUFFER2=null
        overridePropertyWithProxy(AnalyserNode.prototype, 'getFloatFrequencyData',{
            apply: function (target, ctx, args) {
                const results_1 = cache.Reflect.apply(target, ctx, args);
                if (BUFFER2 !== results_1) {
                    BUFFER2 = results_1;
                  for (var i = 0; i < results_1.length; i += 100) {
                    let index = Math.floor(Math.random() * i);
                    results_1[index] = results_1[index] + Math.random() * 0.1;
                  }
                }
                return results_1;
            },
            get: function (target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(...arguments);
            },
        });
    };
    changeCanvasMain()
    changeWebGlMain()
    changeAudio()
}
function getInjectableFingerprintFunction() {
    var mainFunctionString2 = inject2.toString();
    const generator = new fingerprint_generator_1.FingerprintGenerator();
    const fingerprintWithHeaders =  generator.getFingerprint({
            browsers: ["chrome"],
            devices: ["desktop"],
            operatingSystems: ["macos"],
            mockWebRTC: true,
        });
    var mainFunctionString = injector.getInjectableScript(fingerprintWithHeaders);
    mainFunctionString = prettierJs(mainFunctionString)
    mainFunctionString2 = prettierJs(mainFunctionString2)
    // mainFunctionString = mainFunctionString.replaceAll(`"userAgent": `, `// "userAgent": `);
    // mainFunctionString = mainFunctionString.replaceAll(`"languages": `, `// "languages": `);
    // mainFunctionString = mainFunctionString.replaceAll(`"hardwareConcurrency": `, `// "hardwareConcurrency": `);
    return `(()=>{
//start
// mainFunctionString
(${mainFunctionString})();
// mainFunctionString2
(${mainFunctionString2})();
// end
})()`
}

const fs = require('fs');
const data =getInjectableFingerprintFunction()

fs.writeFile('browser/stealthRaw.js', data, (err) => {
  if (err) {
    console.error(err);
    return;
  }
  console.log('文件写入成功！');
});


