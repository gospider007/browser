const { FingerprintInjector } = require('fingerprint-injector');
const { FingerprintGenerator } = require('fingerprint-generator');
const injector = new FingerprintInjector();
const generator = new FingerprintGenerator();

function inject() {
    const { battery, navigator: { 
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    extraProperties, userAgentData, webdriver, ...navigatorProps }, screen: allScreenProps, videoCard, historyLength, audioCodecs, videoCodecs,
    // @ts-expect-error internal browser code
     } = fp;
    const { 
    // window screen props
    outerHeight, outerWidth, devicePixelRatio, innerWidth, innerHeight, screenX, pageXOffset, pageYOffset, 
    // Document screen props
    clientWidth, clientHeight, 
    // Ignore hdr for now.
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    hasHDR, 
    // window.screen props
    ...newScreen } = allScreenProps;
    const windowScreenProps = {
        innerHeight,
        outerHeight,
        outerWidth,
        innerWidth,
        screenX,
        pageXOffset,
        pageYOffset,
        devicePixelRatio,
    };
    const documentScreenProps = {
        clientHeight,
        clientWidth,
    };
    runHeadlessFixes();
    if (userAgentData) {
        overrideUserAgentData(userAgentData);
    }
    if (window.navigator.webdriver) {
        navigatorProps.webdriver = false;
    }
    overrideInstancePrototype(window.navigator, navigatorProps);
    overrideInstancePrototype(window.screen, newScreen);
    overrideWindowDimensionsProps(windowScreenProps);
    overrideDocumentDimensionsProps(documentScreenProps);
    overrideInstancePrototype(window.history, { length: historyLength });
    overrideWebGl(videoCard);
    overrideCodecs(audioCodecs, videoCodecs);
    overrideBattery(battery);
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
function getInjectableFingerprintFunction(fingerprint) {
    const mainFunctionString = inject.toString();
    const mainFunctionString2 = inject2.toString();
    return `(()=>{${injector.utilsJs}; 
const fp=${JSON.stringify(fingerprint,null,2)}; 
(${mainFunctionString})();
(${mainFunctionString2})();
})()`;
}
function createFp(params) {
    //以下这些函数禁止使用, overrideIntlAPI , overrideStatic
    injectable_code=getInjectableFingerprintFunction(params);
    return {result:injectable_code}
}
