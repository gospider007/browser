const { FingerprintInjector } = require('fingerprint-injector');
const fingerprint_generator_1 = require("fingerprint-generator");
const injector = new FingerprintInjector();
const beautify = require("js-beautify").js_beautify;

function prettierJs(code) {
    code = beautify(code, { indent_size: 2 });
    return code
}

function inject2() {
    var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
        if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
            if (ar || !(i in from)) {
                if (!ar) ar = Array.prototype.slice.call(from, 0, i);
                ar[i] = from[i];
            }
        }
        return to.concat(ar || Array.prototype.slice.call(from));
    };

    var seededRandom = function (seed, max, min) {
        if (max === void 0) { max = 1; }
        if (min === void 0) { min = 0; }
        if (typeof seed === 'string') {
            seed = hashNumberFromString(seed);
        }
        var mod = 233280;
        seed = (seed * 9301 + 49297) % mod;
        if (seed < 0)
            seed += mod; // 确保 seed 为正数
        var rnd = seed / mod;
        return min + rnd * (max - min);
    };

    var seededEl = function (arr, seed) {
        return arr[seed % arr.length];
    };

    var shuffleArray = function (array, seed) {
        var _array = __spreadArray([], array, true);
        var m = _array.length, t, i;
        var random = function () {
            var x = Math.sin(seed++) * 10000;
            return x - Math.floor(x);
        };
        while (m) {
            i = Math.floor(random() * m--);
            t = _array[m];
            _array[m] = _array[i];
            _array[i] = t;
        }
        return _array;
    };


    var genRandomSeed = function () {
        return Math.floor(seededRandom(Math.random() * 1e6, Number.MAX_SAFE_INTEGER, 1));
    };

    var hashNumberFromString = function (input) {
        var hash = 0;
        for (var i = 0; i < input.length; i++) {
            var char = input.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash;
        }
        return Math.abs(hash % Number.MAX_SAFE_INTEGER);
    };

    var arrayFilter = function (arr) {
        return arr.filter(function (item) { return item !== undefined && item !== false; });
    };


    var getNextPowerOfTen = function (num) {
        if (num === 0)
            return 0;
        else if (num === 1)
            return 10;
        if (num < 0)
            num = -num;
        return Math.pow(10, Math.ceil(Math.log10(num)));
    };

    /**
     * 对象提取值
     */
    var pick = function (obj, keys) {
        var res = {};
        for (var _i = 0, keys_1 = keys; _i < keys_1.length; _i++) {
            var key = keys_1[_i];
            res[key] = obj[key];
        }
        return res;
    };

    var randomWebglNoise = function (seed) {
        return [seededRandom(seed, 1, -1), seededRandom(seed + 1, 1, -1)];
    };
    /**
     * 获取随机字体噪音
     */
    var randomFontNoise = function (seed, mark) {
        var random = seededRandom((seed + hashNumberFromString(mark)) % Number.MAX_SAFE_INTEGER, 3, 0);
        if ((random * 10) % 1 < 0.9)
            return 0;
        return Math.floor(random) - 1;
    };


    var isPixelEqual = function (p1, p2) {
        return p1[0] === p2[0] && p1[1] === p2[1] && p1[2] === p2[2] && p1[3] === p2[3];
    };
    var pixelCopy = function (src, dst, index) {
        dst[0] = src[index];
        dst[1] = src[index + 1];
        dst[2] = src[index + 2];
        dst[3] = src[index + 3];
    };
    /**
     * 在2d画布绘制噪音
     */
    var drawNoise = function (rawFunc, rawSeed, ctx, sx, sy, sw, sh, settings) {
        var imageData = rawFunc.call(ctx, sx, sy, sw, sh, settings);
        var isChanged = false;
        var Arr = Uint8ClampedArray;
        var center = new Arr(4);
        var up = new Arr(4);
        var down = new Arr(4);
        var left = new Arr(4);
        var right = new Arr(4);
        var pixelData = imageData.data;
        var noiseIndex = 0;

        outer: for (var row = 1; row < sh - 2; row += 2) {
            for (var col = 1; col < sw - 2; col += 2) {
                var index = (row * sw + col) * 4;
                pixelCopy(pixelData, center, index);
                pixelCopy(pixelData, up, ((row - 1) * sw + col) * 4);
                if (isPixelEqual(center, up))
                    continue;
                pixelCopy(pixelData, down, ((row + 1) * sw + col) * 4);
                if (isPixelEqual(center, down))
                    continue;
                pixelCopy(pixelData, left, (row * sw + (col - 1)) * 4);
                if (isPixelEqual(center, left))
                    continue;
                pixelCopy(pixelData, right, (row * sw + (col + 1)) * 4);
                if (isPixelEqual(center, right))
                    continue;
                noiseIndex++;
                // if (noiseIndex > 260) {
                //     break outer;
                // }
                let seed = (rawSeed + noiseIndex) % 256

                let n = Math.floor(seededRandom(seed + noiseIndex, 255, 0)) % 256;
                let r = Math.sin(seed + noiseIndex) * 10000 - Math.floor(Math.sin(seed + noiseIndex) * 10000);
                const noiseStrength = 1 + (seed % 3);              // 1～3 强度

                pixelData[index + 0] = (pixelData[index + 0] + n * r) & 255;
                pixelData[index + 1] = (pixelData[index + 1] ^ n) & 255;
                pixelData[index + 2] = (pixelData[index + 2] + (n >> 1)) & 255;
                pixelData[index + 3] = (pixelData[index + 3] + n * noiseStrength) & 255;
                // pixelData[index + 3] = (Math.floor(pixelData[index + 3] + n * noiseStrength)) % 256;
                // pixelData[index + 3] = (Math.floor(seededRandom(seed + noiseIndex, 255, 0))) % 256;
                isChanged = true;
            }
        }
        if (isChanged) {
            ctx.putImageData(imageData, sx, sy);
        }
        return imageData;
    };
    /**
     * 在webgl上下文绘制噪音点
     * @param noisePosition 区间[-1, 1]
     */
    var drawNoiseToWebgl = function (gl, noisePosition) {
        var vertexShaderSource = "attribute vec4 noise;void main() {gl_Position = noise;gl_PointSize = 0.001;}";
        var fragmentShaderSource = "void main() {gl_FragColor = vec4(0.0, 0.0, 0.0, 0.01);}";
        var createShader = function (gl, type, source) {
            var shader = gl.createShader(type);
            if (!shader)
                return;
            gl.shaderSource(shader, source);
            gl.compileShader(shader);
            return shader;
        };
        var vertexShader = createShader(gl, gl.VERTEX_SHADER, vertexShaderSource);
        var fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, fragmentShaderSource);
        if (!vertexShader || !fragmentShader)
            return;
        var program = gl.createProgram();
        if (!program)
            return;
        gl.attachShader(program, vertexShader);
        gl.attachShader(program, fragmentShader);
        gl.linkProgram(program);
        gl.useProgram(program);
        var positions = new Float32Array(noisePosition);
        var positionBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
        gl.bufferData(gl.ARRAY_BUFFER, positions, gl.STATIC_DRAW);
        var noise = gl.getAttribLocation(program, 'noise');
        gl.enableVertexAttribArray(noise);
        gl.vertexAttribPointer(noise, 2, gl.FLOAT, false, 0, 0);
        gl.drawArrays(gl.POINTS, 0, 1);
    };

    function changeCanvas() {
        overridePropertyWithProxy(HTMLCanvasElement.prototype, 'getContext', {
            apply: function (target, thisArg, args) {
                if (args[0] === '2d') {
                    const option = args[1] ?? {};
                    option.willReadFrequently = true;
                    args[1] = option
                }
                return cache.Reflect.apply(target, thisArg, args);
            }
        });
        overridePropertyWithProxy(CanvasRenderingContext2D.prototype, 'getImageData', {
            apply: function (target, thisArg, args) {
                return drawNoise(target, seed, thisArg, ...args);
            }
        })
    }
    function changeWebgl() {
        var noise = randomWebglNoise(seed);
        var handler = {
            apply: function (target, thisArg, args) {
                drawNoiseToWebgl(thisArg, noise);
                return cache.Reflect.apply(target, thisArg, args);
            }
        };
        overridePropertyWithProxy(WebGLRenderingContext.prototype, 'readPixels', handler);
        overridePropertyWithProxy(WebGL2RenderingContext.prototype, 'readPixels', handler);

        var noise2 = seededRandom(seed, 1, 0);
        var handler = {
            apply: function (target, thisArg, args) {
                const res = target.apply(thisArg, args);
                res?.push?.('EXT_' + noise2);
                return res;
            }
        };
        overridePropertyWithProxy(WebGLRenderingContext.prototype, 'getSupportedExtensions', handler);
        overridePropertyWithProxy(WebGL2RenderingContext.prototype, 'getSupportedExtensions', handler);
    }
    function changeDataURL() {
        var noiseWebgl = randomWebglNoise(seed);
        overridePropertyWithProxy(HTMLCanvasElement.prototype, 'toDataURL', {
            apply: function (target, thisArg, args) {
                /* 2d */

                const ctx = thisArg.getContext('2d');
                if (ctx) {
                    ctx.getImageData(0, 0, thisArg.width, thisArg.height)
                    return cache.Reflect.apply(target, thisArg, args);
                }

                /* webgl */
                const gl = thisArg.getContext('webgl') ?? thisArg.getContext('webgl2')
                if (gl) {
                    drawNoiseToWebgl(gl, noiseWebgl)
                    return cache.Reflect.apply(target, thisArg, args);
                }

                return cache.Reflect.apply(target, thisArg, args);
            }
        });
    }
    /**
     * Audio
     * 音频指纹
     */
    function changeAudio() {
        const mem = new WeakSet()
        overridePropertyWithProxy(AudioBuffer.prototype, 'getChannelData', {
            apply: function (target, thisArg, args) {
                const data = target.apply(thisArg, args)
                if (mem.has(data)) return data;

                const step = data.length > 2000 ? 100 : 20;
                for (let i = 0; i < data.length; i += step) {
                    const v = data[i]
                    if (v !== 0 && Math.abs(v) > 1e-7) {
                        data[i] += seededRandom(seed + i) * 1e-7;
                    }
                }
                mem.add(data)
                return data;
            }
        });
        var copyHand = {
            apply: function (target, thisArg, args) {
                const channel = args[1]
                if (channel != null) {
                    thisArg.getChannelData(channel)
                }
                return target.apply(thisArg, args)
            }
        };
        overridePropertyWithProxy(AudioBuffer.prototype, 'copyFromChannel', copyHand)
        overridePropertyWithProxy(AudioBuffer.prototype, 'copyToChannel', copyHand)
        overrideGetterWithProxy(DynamicsCompressorNode.prototype, 'reduction', {
            apply(target, thisArg, args) {
                const res = cache.Reflect.apply(target, thisArg, args);
                return (typeof res === 'number' && res !== 0) ? res + dcNoise : res;
            }
        })
    }
    /**
       * Font
       * 字体指纹
       */
    function changeFont() {
        var getHandler = (key) => {
            return {
                apply(target, thisArg, args) {
                    const result = cache.Reflect.apply(target, thisArg, args);
                    const mark = (thisArg.style?.fontFamily ?? key) + result;
                    return result + randomFontNoise(seed, mark);
                },
                get(target, prop, receiver) {
                    useStrictModeExceptions(prop);
                    return Reflect.get(target, prop, receiver);
                },
            }
        }

        overrideGetterWithProxy(HTMLElement.prototype, 'offsetHeight', getHandler("offsetHeight"))
        overrideGetterWithProxy(HTMLElement.prototype, 'offsetWidth', getHandler("offsetWidth"))


        overridePropertyWithProxy(window, 'FontFace', {
            construct: function (target, args, newTarget) {
                const source = args[1]
                if (typeof source === 'string' && source.startsWith('local(')) {
                    const name = source.substring(source.indexOf('(') + 1, source.indexOf(')'));
                    const rand = seededRandom(name + seed, 1, 0);
                    if (rand < 0.02) {
                        args[1] = `local("${rand}")`
                    } else if (rand < 0.04) {
                        args[1] = 'local("Arial")'
                    }
                }
                return new target(...args)
            },
        });
    }
    /**
     * Webgpu
     */



    // ===================== 自定义安全 Proxy =====================
    function createSafeProxy(target, handler) {
        const symbolRaw = Symbol('raw'); // 提供访问原对象接口
        const proxy = new Proxy(target, {
            ...handler,
            get(t, prop, receiver) {
                if (prop === symbolRaw) return t;         // 外部可拿到原对象
                if (prop === 'caller' || prop === 'arguments') return t[prop]; // 防严格模式报错
                const originalGet = handler.get ?? Reflect.get;
                return originalGet(t, prop, receiver);
            },
            setPrototypeOf(t, proto) {
                try {
                    return Reflect.setPrototypeOf(t, proto);
                } catch (e) {
                    const stack = e.stack.split('\n');
                    stack.splice(1, 2);
                    e.stack = stack.join('\n');
                    throw e;
                }
            }
        });
        return proxy
    }



    /* GPUAdapter & GPUDevice */
    function changeWebgpu() {
        const makeNoise = (raw, offset) => {
            const rn = seededRandom(seed + (offset * 7), 64, 1)
            return raw ? raw - Math.floor(rn) : raw;
        }
        // ===================== GPU limits 加噪 Handler =====================
        const limitsHandler = {
            apply(target, thisArg, args) {
                // 调用原始 getter 返回 limits 对象
                const limits = Reflect.apply(target, thisArg, args);

                // 返回安全 Proxy，拦截内部属性访问
                return createSafeProxy(limits, {
                    get(target, prop, receiver) {
                        switch (prop) {
                            case 'maxBufferSize':
                                return makeNoise(target[prop], 0);
                            case 'maxStorageBufferBindingSize':
                                return makeNoise(target[prop], 1);
                            default:
                                const value = target[prop];
                                return typeof value === 'function' ? value.bind(target) : value;
                        }
                    }
                });
            },
            get(target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(target, prop, receiver);
            }
        };
        GPUAdapter && overrideGetterWithProxy(GPUAdapter.prototype, 'limits', limitsHandler);
        GPUDevice && overrideGetterWithProxy(GPUDevice.prototype, 'limits', limitsHandler);

        // @ts-ignore
        if (GPUCommandEncoder?.prototype?.beginRenderPass) {
            // @ts-ignore
            overridePropertyWithProxy(GPUCommandEncoder.prototype, "beginRenderPass", {
                apply(target, self, args) {
                    if (args?.[0]?.colorAttachments?.[0]?.clearValue) {
                        try {
                            const _clearValue = args[0].colorAttachments[0].clearValue
                            let offset = 0
                            for (let key in _clearValue) {
                                let value = _clearValue[key]
                                const noise = seededRandom(seed + (offset++ * 7), 0.01, 0.001)
                                value += value * noise * -1
                                _clearValue[key] = Math.abs(value)
                            }
                            args[0].colorAttachments[0].clearValue = _clearValue;
                        } catch (e) { }
                    }
                    return target.apply(self, args);
                }
            })
        }
        if (GPUQueue?.prototype?.writeBuffer) {
            overridePropertyWithProxy(GPUQueue.prototype, "writeBuffer", {
                apply(target, self, args) {
                    const _data = args?.[2]
                    if (_data && _data instanceof Float32Array) {
                        try {
                            const count = Math.ceil(_data.length * 0.05)
                            let offset = 0
                            const selected = Array(_data.length)
                                .map((_, i) => i)
                                .sort(() => seededRandom(seed + (offset++ * 7), 1, -1))
                                .slice(0, count);

                            offset = 0
                            for (let i = 0; i < selected.length; i++) {
                                const index = selected[i];
                                let value = _data[index];
                                const noise = seededRandom(seed + (offset++ * 7), +0.0001, -0.0001)
                                _data[index] += noise * value;
                            }
                            // args[2] = _data;
                        } catch (e) { }
                    }
                    return target.apply(self, args);
                }
            })
        }
    }
    /**
       * DomRect
       */
    function changeDomRect() {
        var mine = new WeakSet();
        var handler = {
            construct: function (target, args, newTarget) {
                var res = Reflect.construct(target, args, newTarget);
                mine.add(res)
                return res;
            }
        }
        overridePropertyWithProxy(window, "DOMRect", handler)
        overridePropertyWithProxy(window, "DOMRectReadOnly", handler)
        const noise = seededRandom(seed, 1e-6, -1e-6);
        var handler2 = {
            apply(target, thisArg, args) {
                const res = cache.Reflect.apply(target, thisArg, args);
                if (mine.has(thisArg)) return res;
                return res + noise;
            },
            get(target, prop, receiver) {
                useStrictModeExceptions(prop);
                return Reflect.get(target, prop, receiver);
            },
        }
        overrideGetterWithProxy(DOMRect.prototype, "x", handler2)
        overrideGetterWithProxy(DOMRect.prototype, "y", handler2)
        overrideGetterWithProxy(DOMRect.prototype, "width", handler2)
        overrideGetterWithProxy(DOMRect.prototype, "height", handler2)

        var getHandler = (toResult) => {
            return {
                apply: function (target, thisArg, args) {
                    return toResult(thisArg);
                },
                get(target, prop, receiver) {
                    useStrictModeExceptions(prop);
                    return Reflect.get(target, prop, receiver);
                },
            }
        }
        overrideGetterWithProxy(DOMRectReadOnly.prototype, "top", getHandler((rect) => { return rect.y; }))
        overrideGetterWithProxy(DOMRectReadOnly.prototype, "left", getHandler((rect) => { return rect.x; }))
        overrideGetterWithProxy(DOMRectReadOnly.prototype, "bottom", getHandler((rect) => { return rect.y + rect.height; }))
        overrideGetterWithProxy(DOMRectReadOnly.prototype, "right", getHandler((rect) => { return rect.x + rect.width; }))

        overridePropertyWithProxy(DOMRectReadOnly.prototype, 'toJSON', {
            apply: function (target, thisArg, args) {
                return pick(thisArg, ['x', 'y', 'width', 'height', 'bottom', 'left', 'right', 'top']);
            }
        });
    }
    function changeOpen() {
        overridePropertyWithProxy(window, 'open', {
            apply(target, thisArg, args) {
                const url = args[0];
                if (typeof url === 'string') {
                    location.assign(url);
                    return window;
                }
                // 极端兜底（url 不是 string）
                return Reflect.apply(target, thisArg, args);
            }
        });
    }
    var seed = (new Date()).valueOf();
    changeCanvas()
    changeWebgl()
    changeDataURL()
    changeAudio()
    changeFont()
    changeWebgpu()
    changeDomRect()
    changeOpen()
}
/**
 * 递归删除对象或数组中：
 * - 值为字符串且包含 valueKeywords 数组中的任意关键字
 * - 或键名包含 keyKeywords 数组中的任意关键字
 *
 * @param {object|array} data - 要处理的数据
 * @param {string|string[]} valueKeywords - 值中匹配的关键字或关键字数组
 * @param {string|string[]} keyKeywords - 键名匹配的关键字或关键字数组
 * @returns {object|array} 处理后的数据（原地修改）
 */
function removeKeysByCondition(data, valueKeywords, keyKeywords) {
    // 转成数组，方便统一处理
    const valueArr = Array.isArray(valueKeywords) ? valueKeywords : [valueKeywords];
    const keyArr = Array.isArray(keyKeywords) ? keyKeywords : [keyKeywords];

    // 构造正则数组
    const valueRegexArr = valueArr.map(k => new RegExp(k, "i"));
    const keyRegexArr = keyArr.map(k => new RegExp(k, "i"));

    // 判断字符串是否匹配关键字数组
    function matchValue(str) {
        return valueRegexArr.some(regex => regex.test(str));
    }

    function matchKey(str) {
        return keyRegexArr.some(regex => regex.test(str));
    }

    // 处理数组
    if (Array.isArray(data)) {
        for (let i = data.length - 1; i >= 0; i--) {
            const item = data[i];

            if (typeof item === "string" && matchValue(item)) {
                data.splice(i, 1);
            } else if (typeof item === "object" && item !== null) {
                removeKeysByCondition(item, valueKeywords, keyKeywords);
                if (Object.keys(item).length === 0 && !Array.isArray(item)) {
                    data.splice(i, 1);
                }
            }
        }
    }
    // 处理对象
    else if (typeof data === "object" && data !== null) {
        for (const key in data) {
            const value = data[key];

            if (matchKey(key)) {
                delete data[key];
                continue;
            }

            if (typeof value === "string" && matchValue(value)) {
                delete data[key];
            } else if (typeof value === "object" && value !== null) {
                removeKeysByCondition(value, valueKeywords, keyKeywords);
                if (Object.keys(value).length === 0) {
                    delete data[key];
                }
            }
        }
    }
    return data;
}


function getInjectableFingerprintFunction() {
    var mainFunctionString2 = inject2.toString();
    const generator = new fingerprint_generator_1.FingerprintGenerator();
    var fingerprintWithHeaders = generator.getFingerprint({
        browsers: [
            { name: "chrome", minVersion: 139, maxVersion: 139, httpVersion: "2" },
        ],
        devices: ["desktop"],
        operatingSystems: ["macos"],
        locales: ["en-US", "en"],
        slim: false,
        strict: true,
    });
    // fingerprintWithHeaders = removeKeysByCondition(fingerprintWithHeaders, [], ["^userAgentData$"])
    // fingerprintWithHeaders = removeKeysByCondition(fingerprintWithHeaders, [], ["^videoCard$"])
    // fingerprintWithHeaders = removeKeysByCondition(fingerprintWithHeaders, [], ["^language$"])
    // fingerprintWithHeaders = removeKeysByCondition(fingerprintWithHeaders, [], ["^userAgent$", "^appVersion$"])
    var mainFunctionString = injector.getInjectableScript(fingerprintWithHeaders);
    console.log(fingerprintWithHeaders)
    mainFunctionString = prettierJs(mainFunctionString)
    mainFunctionString = mainFunctionString.replaceAll("})()", "})();");
    mainFunctionString = mainFunctionString.replaceAll("\n})();", `
        (${mainFunctionString2})();` + "\n})();");
    mainFunctionString = prettierJs(mainFunctionString)
    mainFunctionString = mainFunctionString.replaceAll(`    overrideUserAgentData(userAgentData)`, `// overrideUserAgentData(userAgentData)`);
    // mainFunctionString = mainFunctionString.replaceAll(`    overrideIntlAPI(navigatorProps.language)`, `// overrideIntlAPI(navigatorProps.language)`);
    mainFunctionString = mainFunctionString.replaceAll(`    overrideWebGl(videoCard);`, `// overrideWebGl(videoCard);`);
    // mainFunctionString = mainFunctionString.replaceAll(`    overrideInstancePrototype(window.navigator, navigatorProps);`, `// overrideInstancePrototype(window.navigator, navigatorProps);`);
    mainFunctionString = mainFunctionString.replaceAll(`    overrideInstancePrototype(window.navigator, navigatorProps);`, `    overrideInstancePrototype(window.navigator, {
      "platform": "MacIntel",
      "deviceMemory": 8,
      "hardwareConcurrency": 4,
      "maxTouchPoints": 0,
      "product": "Gecko",
      "productSub": "20030107",
      "vendor": "Google Inc.",
      "vendorSub": null,
      "doNotTrack": null,
      "appCodeName": "Mozilla",
      "appName": "Netscape",
      "oscpu": null,
    });`);
    return mainFunctionString
}

const fs = require('fs');
const data = getInjectableFingerprintFunction()

fs.writeFile('browser/stealthRaw.js', data, (err) => {
    if (err) {
        console.error(err);
        return;
    }
    console.log('文件写入成功！');
});


