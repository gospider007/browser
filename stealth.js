function setDefaultAttr(key,name,l){
  Object.defineProperty(key, "length", {
    "value": l
  });
  Object.defineProperty(key, "toString", {
      "value": () => `function ${name}() { [native code] }`
  });

  Object.defineProperty(key, "name", {
      "value": name
  });
}
function changeCanvasMain(){
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
  let ctxInf = [];

  const rawGetImageData = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, 'getImageData');
  let noisify = function (canvas, context) {
    if (!context){
    return
    }
    let ctxIdx = ctxArr.indexOf(context);
    let info = ctxInf[ctxIdx];
    const width = canvas.width, height = canvas.height;
    const imageData = rawGetImageData.value.apply(context, [0, 0, width, height]);
    if (info.useArc || info.useFillText) {
        for (let i = 0; i < height; i++) {
            for (let j = 0; j < width; j++) {
                const n = ((i * (width * 4)) + (j * 4));
                imageData.data[n + 0] = imageData.data[n + 0] + rsalt;
                imageData.data[n + 1] = imageData.data[n + 1] + gsalt;
                imageData.data[n + 2] = imageData.data[n + 2] + bsalt;
                imageData.data[n + 3] = imageData.data[n + 3] + asalt;
            }
        }
    }
    context.putImageData(imageData, 0, 0);
  };

  Object.defineProperty(CanvasRenderingContext2D.prototype, "getImageData", {
  ...rawGetImageData,
    "value": function () {
        noisify(this.canvas, this);
        return rawGetImageData.value.apply(this, arguments);
    }
  });
  setDefaultAttr(CanvasRenderingContext2D.prototype.getImageData,"getImageData",4)

  const rawGetContext = Object.getOwnPropertyDescriptor(HTMLCanvasElement.prototype, 'getContext');
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    ...rawGetContext,
    "value": function () {
        let result = rawGetContext.value.apply(this, arguments);
        if (arguments[0] === '2d' && result) {
            ctxArr.push(result)
            ctxInf.push({})
        }
        return result;
    }
  });
  setDefaultAttr(HTMLCanvasElement.prototype.getContext,"getContext",1)

  const rawArc = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, 'arc');
  Object.defineProperty(CanvasRenderingContext2D.prototype, "arc", {
  ...rawArc,
    "value": function () {
        let ctxIdx = ctxArr.indexOf(this);
        ctxInf[ctxIdx].useArc = true;
        return rawArc.value.apply(this, arguments);
    }
  });
  setDefaultAttr(CanvasRenderingContext2D.prototype.arc,"arc",5)

  const rawFillText = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, 'fillText');
  Object.defineProperty(CanvasRenderingContext2D.prototype, "fillText", {
  ...rawFillText,
    "value": function () {
        let ctxIdx = ctxArr.indexOf(this);
        ctxInf[ctxIdx].useFillText = true;
        return rawFillText.value.apply(this, arguments);
    }
  });
  setDefaultAttr(CanvasRenderingContext2D.prototype.fillText,"fillText",3)


  const toBlob = Object.getOwnPropertyDescriptor(HTMLCanvasElement.prototype, 'toBlob');
  Object.defineProperty(HTMLCanvasElement.prototype, "toBlob", {
  ...toBlob,
    "value": function () {
        noisify(this, this.getContext("2d"));
        return toBlob.value.apply(this, arguments);
    }
  });
  setDefaultAttr(HTMLCanvasElement.prototype.toBlob,"toBlob",1)

  const toDataURL = Object.getOwnPropertyDescriptor(HTMLCanvasElement.prototype, 'toDataURL');
  Object.defineProperty(HTMLCanvasElement.prototype, "toDataURL", {
  ...toDataURL,
    "value": function () {
        noisify(this, this.getContext("2d"));
        return toDataURL.value.apply(this, arguments);
    }
  });
  setDefaultAttr(HTMLCanvasElement.prototype.toDataURL,"toDataURL",0)      

};
function changeWebGlMain() {
  var config = {
    "random": {
        "value": function () {
        return Math.random();
        },
        "item": function (e) {
        var rand = e.length * config.random.value();
        return e[Math.floor(rand)];
        },
        "number": function (power) {
        var tmp = [];
        for (var i = 0; i < power.length; i++) {
            tmp.push(Math.pow(2, power[i]));
        }
        /*  */
        return config.random.item(tmp);
        },
        "int": function (power) {
        var tmp = [];
        for (var i = 0; i < power.length; i++) {
            var n = Math.pow(2, power[i]);
            tmp.push(new Int32Array([n, n]));
        }
        /*  */
        return config.random.item(tmp);
        },
        "float": function (power) {
        var tmp = [];
        for (var i = 0; i < power.length; i++) {
            var n = Math.pow(2, power[i]);
            tmp.push(new Float32Array([1, n]));
        }
        /*  */
        return config.random.item(tmp);
        }
    },
    "spoof": {
        "webgl": {
          "buffer": function (target) {
              const bufferData = Object.getOwnPropertyDescriptor(target.prototype, 'bufferData');
              Object.defineProperty(target.prototype, "bufferData", {
              ...bufferData,
              "value": function () {
                  var index = Math.floor(config.random.value() * arguments[1].length);
                  var noise = arguments[1][index] !== undefined ? 0.1 * config.random.value() * arguments[1][index] : 0;
                  arguments[1][index] = arguments[1][index] + noise;
                  return bufferData.value.apply(this, arguments);
              }
              });
              setDefaultAttr(target.prototype.bufferData,"bufferData",3)
          },
        "parameter": function (target) {
            const getParameter = Object.getOwnPropertyDescriptor(target.prototype, 'getParameter');
            Object.defineProperty(target.prototype, "getParameter", {
              ...getParameter,
              "value": function () {
                  if (arguments[0] === 3415) return 0;
                  else if (arguments[0] === 3414) return 24;
                  else if (arguments[0] === 36348) return 30;
                  else if (arguments[0] === 7936) return "WebKit";
                  else if (arguments[0] === 37445) return "Google Inc. (Google)";
                  else if (arguments[0] === 7937) return "WebKit WebGL";
                  else if (arguments[0] === 3379) return config.random.number([14, 15]);
                  else if (arguments[0] === 36347) return config.random.number([12, 13]);
                  else if (arguments[0] === 34076) return config.random.number([14, 15]);
                  else if (arguments[0] === 34024) return config.random.number([14, 15]);
                  else if (arguments[0] === 3386) return config.random.int([13, 14, 15]);
                  else if (arguments[0] === 3413) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 3412) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 3411) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 3410) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 34047) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 34930) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 34921) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 35660) return config.random.number([1, 2, 3, 4]);
                  else if (arguments[0] === 35661) return config.random.number([4, 5, 6, 7, 8]);
                  else if (arguments[0] === 36349) return config.random.number([10, 11, 12, 13]);
                  else if (arguments[0] === 33902) return config.random.float([0, 10, 11, 12, 13]);
                  else if (arguments[0] === 33901) return config.random.float([0, 10, 11, 12, 13]);
                  else if (arguments[0] === 37446) return config.random.item(["Graphics", "HD Graphics", "Intel(R) HD Graphics"]);
                  else if (arguments[0] === 7938) return config.random.item(["WebGL 1.0", "WebGL 1.0 (OpenGL)", "WebGL 1.0 (OpenGL Chromium)"]);
                  else if (arguments[0] === 35724) return config.random.item(["WebGL", "WebGL GLSL", "WebGL GLSL ES", "WebGL GLSL ES (OpenGL Chromium"]);
                  //
                  return getParameter.value.apply(this, arguments);
              }
            });
            setDefaultAttr(target.prototype.getParameter,"getParameter",1)
          }
      }
    }
  };
  config.spoof.webgl.buffer(WebGLRenderingContext);
  config.spoof.webgl.buffer(WebGL2RenderingContext);
  config.spoof.webgl.parameter(WebGLRenderingContext);
  config.spoof.webgl.parameter(WebGL2RenderingContext);
};
function changeMime(){
  const platform = Object.getOwnPropertyDescriptor(Navigator.prototype, 'platform');
  Object.defineProperty(Navigator.prototype, 'platform', {
      ...platform,
      get: function () {
          return 'Win32'
      },
  });
  const hardwareConcurrency = Object.getOwnPropertyDescriptor(Navigator.prototype, 'hardwareConcurrency');
  Object.defineProperty(Navigator.prototype, 'hardwareConcurrency', {
      ...hardwareConcurrency,
      get: function () {
          return 8
      },
  });
};
function changeAudio(){
  const context = {
      "BUFFER": null,
      "getChannelData": function (e) {
        const getChannelData = Object.getOwnPropertyDescriptor(e.prototype, 'getChannelData');
        Object.defineProperty(e.prototype, "getChannelData", {
            ...getChannelData,
            "value": function () {
            const results_1 = getChannelData.value.apply(this, arguments);
            if (context.BUFFER !== results_1) {
                context.BUFFER = results_1;
                for (var i = 0; i < results_1.length; i += 100) {
                let index = Math.floor(Math.random() * i);
                results_1[index] = results_1[index] + Math.random() * 0.0000001;
                }
            }
            return results_1;
            }
        });
        setDefaultAttr(e.prototype.getChannelData,"getChannelData",1)
      },
      "createAnalyser": function (e) {
        const createAnalyser = Object.getOwnPropertyDescriptor(e.prototype.__proto__, 'createAnalyser');
        Object.defineProperty(e.prototype.__proto__, "createAnalyser", {
            ...createAnalyser,
            "value": function () {
              const results_2 = createAnalyser.value.apply(this, arguments);
              const getFloatFrequencyData = Object.getOwnPropertyDescriptor(results_2.__proto__, 'getFloatFrequencyData');
              Object.defineProperty(results_2.__proto__, "getFloatFrequencyData", {
                  ...getFloatFrequencyData,
                  "value": function () {
                    const results_3 = getFloatFrequencyData.value.apply(this, arguments);
                    for (var i = 0; i < arguments[0].length; i += 100) {
                        let index = Math.floor(Math.random() * i);
                        arguments[0][index] = arguments[0][index] + Math.random() * 0.1;
                    }
                    return results_3;
                  }
              });
              setDefaultAttr(results_2.__proto__.getFloatFrequencyData,"getFloatFrequencyData",0)
              return results_2;
            }
        });
        setDefaultAttr(e.prototype.__proto__.createAnalyser,"createAnalyser",0)
      }
  };
  context.getChannelData(AudioBuffer);
  context.createAnalyser(AudioContext);
  context.getChannelData(OfflineAudioContext);
  context.createAnalyser(OfflineAudioContext);
};
function changeFont(){
  var rand = {
      "noise": function () {
      var SIGN = Math.random() < Math.random() ? -1 : 1;
      return Math.floor(Math.random() + SIGN * Math.random());
      },
      "sign": function () {
      const tmp = [-1, -1, -1, -1, -1, -1, +1, -1, -1, -1];
      const index = Math.floor(Math.random() * tmp.length);
      return tmp[index];
      }
  };
  //
  const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight');

  Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
      ...offsetHeight,
      get () {
        const height = Math.floor(this.getBoundingClientRect().height);
        const valid = height && rand.sign() === 1;
        const result = valid ? height + rand.noise() : height;
        return result;
      }
  });
  const offsetWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
      ...offsetWidth,
      get () {
        const width = Math.floor(this.getBoundingClientRect().width);
        const valid = width && rand.sign() === 1;
        const result = valid ? width + rand.noise() : width;
        return result;
      }
  });
};
function changeRect(){
  const height = Object.getOwnPropertyDescriptor(DOMRect.prototype, 'height');

  Object.defineProperty(DOMRect.prototype, "height", {
  ...height,
  get() {
        return this.toJSON()["height"]+Math.random();
    },
  });
  const width = Object.getOwnPropertyDescriptor(DOMRect.prototype, 'width');
  Object.defineProperty(DOMRect.prototype, "width", {
      ...width,
    get() {
        return this.toJSON()["width"]+Math.random();
    },
  });
}
changeCanvasMain()
changeWebGlMain()
changeMime()
changeAudio()
changeFont()
changeRect()

  
  
  