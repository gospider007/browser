(() => {
  "use strict";
  /* eslint-disable no-unused-vars */
  const isHeadlessChromium = /headless/i.test(navigator.userAgent) && navigator.plugins.length === 0;
  const isChrome = navigator.userAgent.includes('Chrome');
  const isFirefox = navigator.userAgent.includes('Firefox');
  const isSafari = navigator.userAgent.includes('Safari') &&
    !navigator.userAgent.includes('Chrome');
  let slim = null;

  function getSlim() {
    if (slim === null) {
      slim = window.slim || false;
      if (typeof window.slim !== 'undefined') {
        delete window.slim;
      }
    }
    return slim;
  }
  // This file contains utils that are build and included on the window object with some randomized prefix.
  // some protections can mess with these to prevent the overrides - our script is first so we can reference the old values.
  const cache = {
    Reflect: {
      get: Reflect.get.bind(Reflect),
      apply: Reflect.apply.bind(Reflect),
    },
    // Used in `makeNativeString`
    nativeToStringStr: `${Function.toString}`, // => `function toString() { [native code] }`
  };
  /**
   * @param masterObject Object to override.
   * @param propertyName Property to override.
   * @param proxyHandler Proxy handled with the new value.
   */
  function overridePropertyWithProxy(masterObject, propertyName, proxyHandler) {
    const originalObject = masterObject[propertyName];
    const proxy = new Proxy(masterObject[propertyName], stripProxyFromErrors(proxyHandler));
    redefineProperty(masterObject, propertyName, {
      value: proxy
    });
    redirectToString(proxy, originalObject);
  }
  const prototypeProxyHandler = {
    setPrototypeOf: (target, newProto) => {
      try {
        throw new TypeError('Cyclic __proto__ value');
      } catch (e) {
        const oldStack = e.stack;
        const oldProto = Object.getPrototypeOf(target);
        Object.setPrototypeOf(target, newProto);
        try {
          // shouldn't throw if prototype is okay, will throw if there is a prototype cycle (maximum call stack size exceeded).
          // eslint-disable-next-line no-unused-expressions
          target.nonexistentpropertytest;
          return true;
        } catch (err) {
          Object.setPrototypeOf(target, oldProto);
          if (oldStack.includes('Reflect.setPrototypeOf'))
            return false;
          const newError = new TypeError('Cyclic __proto__ value');
          const stack = oldStack.split('\n');
          newError.stack = [stack[0], ...stack.slice(2)].join('\n');
          throw newError;
        }
      }
    },
  };

  function useStrictModeExceptions(prop) {
    if (['caller', 'callee', 'arguments'].includes(prop)) {
      throw TypeError(`'caller', 'callee', and 'arguments' properties may not be accessed on strict mode functions or the arguments objects for calls to them`);
    }
  }
  /**
   * @param masterObject Object to override.
   * @param propertyName Property to override.
   * @param proxyHandler ES6 Proxy handler object with a get handle only.
   */
  function overrideGetterWithProxy(masterObject, propertyName, proxyHandler) {
    const fn = Object.getOwnPropertyDescriptor(masterObject, propertyName).get;
    const fnStr = fn.toString; // special getter function string
    const proxyObj = new Proxy(fn, {
      ...stripProxyFromErrors(proxyHandler),
      ...prototypeProxyHandler,
    });
    redefineProperty(masterObject, propertyName, {
      get: proxyObj
    });
    redirectToString(proxyObj, fnStr);
  }
  /**
   * @param instance Instance to override.
   * @param overrideObj New instance values.
   */
  function overrideInstancePrototype(instance, overrideObj) {
    try {
      Object.keys(overrideObj).forEach((key) => {
        if (!(overrideObj[key] === null)) {
          try {
            overrideGetterWithProxy(Object.getPrototypeOf(instance), key, makeHandler().getterValue(overrideObj[key]));
          } catch (e) {
            console.debug(e);
            // console.error(`Could not override property: ${key} on ${instance}. Reason: ${e.message} `); // some fingerprinting services can be listening
          }
        }
      });
    } catch (e) {
      console.error(e);
    }
  }
  /**
   * Updates the .toString method in Function.prototype to return a native string representation of the function.
   * @param {*} proxyObj
   * @param {*} originalObj
   */
  function redirectToString(proxyObj, originalObj) {
    if (getSlim())
      return;
    const handler = {
      setPrototypeOf: (target, newProto) => {
        try {
          throw new TypeError('Cyclic __proto__ value');
        } catch (e) {
          if (e.stack.includes('Reflect.setPrototypeOf'))
            return false;
          // const stack = e.stack.split('\n');
          // e.stack = [stack[0], ...stack.slice(2)].join('\n');
          throw e;
        }
      },
      apply(target, ctx) {
        // This fixes e.g. `HTMLMediaElement.prototype.canPlayType.toString + ""`
        if (ctx === Function.prototype.toString) {
          return makeNativeString('toString');
        }
        // `toString` targeted at our proxied Object detected
        if (ctx === proxyObj) {
          // Return the toString representation of our original object if possible
          return makeNativeString(proxyObj.name);
        }
        // Check if the toString prototype of the context is the same as the global prototype,
        // if not indicates that we are doing a check across different windows., e.g. the iframeWithdirect` test case
        const hasSameProto = Object.getPrototypeOf(Function.prototype.toString).isPrototypeOf(ctx.toString); // eslint-disable-line no-prototype-builtins
        if (!hasSameProto) {
          // Pass the call on to the local Function.prototype.toString instead
          return ctx.toString();
        }
        if (Object.getPrototypeOf(ctx) === proxyObj) {
          try {
            return target.call(ctx);
          } catch (err) {
            err.stack = err.stack.replace('at Object.toString (', 'at Function.toString (');
            throw err;
          }
        }
        return target.call(ctx);
      },
      get(target, prop, receiver) {
        if (prop === 'toString') {
          return new Proxy(target.toString, {
            apply(tget, thisArg, argumentsList) {
              try {
                return tget.bind(thisArg)(...argumentsList);
              } catch (err) {
                if (Object.getPrototypeOf(thisArg) === tget) {
                  err.stack = err.stack.replace('at Object.toString (', 'at Function.toString (');
                }
                throw err;
              }
            },
          });
        }
        useStrictModeExceptions(prop);
        return Reflect.get(target, prop, receiver);
      },
    };
    const toStringProxy = new Proxy(Function.prototype.toString, stripProxyFromErrors(handler));
    redefineProperty(Function.prototype, 'toString', {
      value: toStringProxy,
    });
  }

  function makeNativeString(name = '') {
    return cache.nativeToStringStr.replace('toString', name || '');
  }

  function redefineProperty(masterObject, propertyName, descriptorOverrides = {}) {
    return Object.defineProperty(masterObject, propertyName, {
      // Copy over the existing descriptors (writable, enumerable, configurable, etc)
      ...(Object.getOwnPropertyDescriptor(masterObject, propertyName) || {}),
      // Add our overrides (e.g. value, get())
      ...descriptorOverrides,
    });
  }
  /**
   * For all the traps in the passed proxy handler, we wrap them in a try/catch and modify the error stack if they throw.
   * @param {*} handler A proxy handler object
   * @returns A new proxy handler object with error stack modifications
   */
  function stripProxyFromErrors(handler) {
    const newHandler = {};
    // We wrap each trap in the handler in a try/catch and modify the error stack if they throw
    const traps = Object.getOwnPropertyNames(handler);
    traps.forEach((trap) => {
      newHandler[trap] = function () {
        try {
          // Forward the call to the defined proxy handler
          return handler[trap].apply(this, arguments || []); //eslint-disable-line
        } catch (err) {
          // Stack traces differ per browser, we only support chromium based ones currently
          if (!err || !err.stack || !err.stack.includes(`at `)) {
            throw err;
          }
          // When something throws within one of our traps the Proxy will show up in error stacks
          // An earlier implementation of this code would simply strip lines with a blacklist,
          // but it makes sense to be more surgical here and only remove lines related to our Proxy.
          // We try to use a known "anchor" line for that and strip it with everything above it.
          // If the anchor line cannot be found for some reason we fall back to our blacklist approach.
          const stripWithBlacklist = (stack, stripFirstLine = true) => {
            const blacklist = [
              `at Reflect.${trap} `, // e.g. Reflect.get or Reflect.apply
              `at Object.${trap} `, // e.g. Object.get or Object.apply
              `at Object.newHandler.<computed> [as ${trap}] `, // caused by this very wrapper :-)
              `at newHandler.<computed> [as ${trap}] `, // also caused by this wrapper :p
            ];
            return (err.stack
              .split('\n')
              // Always remove the first (file) line in the stack (guaranteed to be our proxy)
              .filter((line, index) => !(index === 1 && stripFirstLine))
              // Check if the line starts with one of our blacklisted strings
              .filter((line) => !blacklist.some((bl) => line.trim().startsWith(bl)))
              .join('\n'));
          };
          const stripWithAnchor = (stack, anchor) => {
            const stackArr = stack.split('\n');
            // eslint-disable-next-line no-param-reassign
            anchor =
              anchor ||
              `at Object.newHandler.<computed> [as ${trap}] `; // Known first Proxy line in chromium
            const anchorIndex = stackArr.findIndex((line) => line.trim().startsWith(anchor));
            if (anchorIndex === -1) {
              return false; // 404, anchor not found
            }
            // Strip everything from the top until we reach the anchor line
            // Note: We're keeping the 1st line (zero index) as it's unrelated (e.g. `TypeError`)
            stackArr.splice(1, anchorIndex);
            return stackArr.join('\n');
          };
          const oldStackLines = err.stack.split('\n');
          Error.captureStackTrace(err);
          const newStackLines = err.stack.split('\n');
          err.stack = [
            newStackLines[0],
            oldStackLines[1],
            ...newStackLines.slice(1),
          ].join('\n');
          if ((err.stack || '').includes('toString (')) {
            err.stack = stripWithBlacklist(err.stack, false);
            throw err;
          }
          // Try using the anchor method, fallback to blacklist if necessary
          err.stack =
            stripWithAnchor(err.stack) || stripWithBlacklist(err.stack);
          throw err; // Re-throw our now sanitized error
        }
      };
    });
    return newHandler;
  }

  function overrideWebGl(webGl) {
    // try to override WebGl
    try {
      const getParameterProxyHandler = {
        apply(target, ctx, args) {
          const param = (args || [])[0];
          const result = cache.Reflect.apply(target, ctx, args);
          const debugInfo = ctx.getExtension('WEBGL_debug_renderer_info');
          const UNMASKED_VENDOR_WEBGL = (debugInfo && debugInfo.UNMASKED_VENDOR_WEBGL) || 37445;
          const UNMASKED_RENDERER_WEBGL = (debugInfo && debugInfo.UNMASKED_RENDERER_WEBGL) || 37446;
          if (param === UNMASKED_VENDOR_WEBGL) {
            return webGl.vendor;
          }
          if (param === UNMASKED_RENDERER_WEBGL) {
            return webGl.renderer;
          }
          return result;
        },
        get(target, prop, receiver) {
          useStrictModeExceptions(prop);
          return Reflect.get(target, prop, receiver);
        },
      };
      const addProxy = (obj, propName) => {
        overridePropertyWithProxy(obj, propName, getParameterProxyHandler);
      };
      // For whatever weird reason loops don't play nice with Object.defineProperty, here's the next best thing:
      addProxy(WebGLRenderingContext.prototype, 'getParameter');
      addProxy(WebGL2RenderingContext.prototype, 'getParameter');
    } catch (err) {
      console.warn(err);
    }
  }
  const overrideCodecs = (audioCodecs, videoCodecs) => {
    try {
      const codecs = {
        ...Object.fromEntries(Object.entries(audioCodecs).map(([key, value]) => [
          `audio/${key}`,
          value,
        ])),
        ...Object.fromEntries(Object.entries(videoCodecs).map(([key, value]) => [
          `video/${key}`,
          value,
        ])),
      };
      const findCodec = (codecString) => {
        const [mime, codecSpec] = codecString.split(';');
        if (mime === 'video/mp4') {
          if (codecSpec && codecSpec.includes('avc1.42E01E')) {
            // codec is missing from Chromium
            return {
              name: mime,
              state: 'probably'
            };
          }
        }
        const codec = Object.entries(codecs).find(([key]) => key === codecString.split(';')[0]);
        if (codec) {
          return {
            name: codec[0],
            state: codec[1]
          };
        }
        return undefined;
      };
      const canPlayType = {
        // eslint-disable-next-line
        apply: function (target, ctx, args) {
          if (!args || !args.length) {
            return target.apply(ctx, args);
          }
          const [codecString] = args;
          const codec = findCodec(codecString);
          if (codec) {
            return codec.state;
          }
          // If the codec is not in our collected data use
          return target.apply(ctx, args);
        },
      };
      overridePropertyWithProxy(HTMLMediaElement.prototype, 'canPlayType', canPlayType);
    } catch (e) {
      console.warn(e);
    }
  };

  function overrideBattery(batteryInfo) {
    try {
      const getBattery = {
        ...prototypeProxyHandler,
        // eslint-disable-next-line
        apply: async function () {
          return batteryInfo;
        },
      };
      if (navigator.getBattery) {
        // Firefox does not have this method - to be fixed
        overridePropertyWithProxy(Object.getPrototypeOf(navigator), 'getBattery', getBattery);
      }
    } catch (e) {
      console.warn(e);
    }
  }

  function overrideIntlAPI(language) {
    try {
      const innerHandler = {
        construct(Target, [locales, options]) {
          return new Target(locales ?? language, options);
        },
        apply(target, _, [locales, options]) {
          return target(locales ?? language, options);
        },
      };
      overridePropertyWithProxy(window, 'Intl', {
        get(target, key) {
          if (typeof key !== 'string' || key[0].toLowerCase() === key[0])
            return target[key];
          return new Proxy(target[key], innerHandler);
        },
      });
    } catch (e) {
      console.warn(e);
    }
  }

  function makeHandler() {
    return {
      // Used by simple `navigator` getter evasions
      getterValue: (value) => ({
        apply(target, ctx, args) {
          // Let's fetch the value first, to trigger and escalate potential errors
          // Illegal invocations like `navigator.__proto__.vendor` will throw here
          const ret = cache.Reflect.apply(...arguments); // eslint-disable-line
          if (args && args.length === 0) {
            return value;
          }
          return ret;
        },
        get(target, prop, receiver) {
          useStrictModeExceptions(prop);
          return Reflect.get(target, prop, receiver);
        },
      }),
    };
  }

  function overrideScreenByReassigning(target, newProperties) {
    for (const [prop, value] of Object.entries(newProperties)) {
      if (value > 0) {
        // The 0 values are introduced by collecting in the hidden iframe.
        // They are document sizes anyway so no need to test them or inject them.
        // eslint-disable-next-line no-param-reassign
        target[prop] = value;
      }
    }
  }

  function overrideWindowDimensionsProps(props) {
    try {
      overrideScreenByReassigning(window, props);
    } catch (e) {
      console.warn(e);
    }
  }

  function overrideDocumentDimensionsProps(props) {
    try {
      // FIX THIS = non-zero values here block the injecting process?
      // overrideScreenByReassigning(window.document.body, props);
    } catch (e) {
      console.warn(e);
    }
  }

  function replace(target, key, value) {
    if (target?.[key]) {
      // eslint-disable-next-line no-param-reassign
      target[key] = value;
    }
  }
  // Replaces all the WebRTC related methods with a recursive ES6 Proxy
  // This way, we don't have to model a mock WebRTC API and we still don't get any exceptions.
  function blockWebRTC() {
    const handler = {
      get: () => {
        return new Proxy(() => { }, handler);
      },
      apply: () => {
        return new Proxy(() => { }, handler);
      },
      construct: () => {
        return new Proxy(() => { }, handler);
      },
    };
    const ConstrProxy = new Proxy(Object, handler);
    const proxy = new Proxy(() => { }, handler);
    replace(navigator.mediaDevices, 'getUserMedia', proxy);
    replace(navigator, 'webkitGetUserMedia', proxy);
    replace(navigator, 'mozGetUserMedia', proxy);
    replace(navigator, 'getUserMedia`', proxy);
    replace(window, 'webkitRTCPeerConnection', proxy);
    replace(window, 'RTCPeerConnection', ConstrProxy);
    replace(window, 'MediaStreamTrack', ConstrProxy);
  }

  function overrideUserAgentData(userAgentData) {
    try {
      const {
        brands,
        mobile,
        platform,
        ...highEntropyValues
      } = userAgentData;
      // Override basic properties
      const getHighEntropyValues = {
        // eslint-disable-next-line
        apply: async function (target, ctx, args) {
          // Just to throw original validation error
          // Remove traces of our Proxy
          const stripErrorStack = (stack) => stack
            .split('\n')
            .filter((line) => !line.includes('at Object.apply'))
            .filter((line) => !line.includes('at Object.get'))
            .join('\n');
          try {
            if (!args || !args.length) {
              return target.apply(ctx, args);
            }
            const [hints] = args;
            await target.apply(ctx, args);
            const data = {
              brands,
              mobile,
              platform
            };
            hints.forEach((hint) => {
              data[hint] = highEntropyValues[hint];
            });
            return data;
          } catch (err) {
            err.stack = stripErrorStack(err.stack);
            throw err;
          }
        },
      };
      if (window.navigator.userAgentData) {
        // Firefox does not contain this property - to be fixed
        overridePropertyWithProxy(Object.getPrototypeOf(window.navigator.userAgentData), 'getHighEntropyValues', getHighEntropyValues);
        overrideInstancePrototype(window.navigator.userAgentData, {
          brands,
          mobile,
          platform,
        });
      }
    } catch (e) {
      console.warn(e);
    }
  }

  function fixWindowChrome() {
    if (isChrome && !window.chrome) {
      Object.defineProperty(window, 'chrome', {
        writable: true,
        enumerable: true,
        configurable: false,
        value: {}, // incomplete, todo!
      });
    }
  }
  // heavily inspired by https://github.com/berstend/puppeteer-extra/, check it out!
  function fixPermissions() {
    const isSecure = document.location.protocol.startsWith('https');
    if (isSecure) {
      overrideGetterWithProxy(Notification, 'permission', {
        apply() {
          return 'default';
        },
      });
    }
    if (!isSecure) {
      const handler = {
        apply(target, ctx, args) {
          const param = (args || [])[0];
          const isNotifications = param && param.name && param.name === 'notifications';
          if (!isNotifications) {
            return cache.Reflect.apply(target, ctx, args);
          }
          return Promise.resolve(Object.setPrototypeOf({
            state: 'denied',
            onchange: null,
          }, PermissionStatus.prototype));
        },
      };
      overridePropertyWithProxy(Permissions.prototype, 'query', handler);
    }
  }

  function fixIframeContentWindow() {
    try {
      // Adds a contentWindow proxy to the provided iframe element
      const addContentWindowProxy = (iframe) => {
        const contentWindowProxy = {
          get(target, key) {
            if (key === 'self') {
              return this;
            }
            if (key === 'frameElement') {
              return iframe;
            }
            if (key === '0') {
              return undefined;
            }
            return Reflect.get(target, key);
          },
        };
        if (!iframe.contentWindow) {
          const proxy = new Proxy(window, contentWindowProxy);
          Object.defineProperty(iframe, 'contentWindow', {
            get() {
              return proxy;
            },
            set(newValue) { },
            enumerable: true,
            configurable: false,
          });
        }
      };
      // Handles iframe element creation, augments `srcdoc` property so we can intercept further
      const handleIframeCreation = (target, thisArg, args) => {
        const iframe = target.apply(thisArg, args);
        // We need to keep the originals around
        const _iframe = iframe;
        const _srcdoc = _iframe.srcdoc;
        // Add hook for the srcdoc property
        // We need to be very surgical here to not break other iframes by accident
        Object.defineProperty(iframe, 'srcdoc', {
          configurable: true, // Important, so we can reset this later
          get() {
            return _srcdoc;
          },
          set(newValue) {
            addContentWindowProxy(this);
            // Reset property, the hook is only needed once
            Object.defineProperty(iframe, 'srcdoc', {
              configurable: false,
              writable: false,
              value: _srcdoc,
            });
            _iframe.srcdoc = newValue;
          },
        });
        return iframe;
      };
      // Adds a hook to intercept iframe creation events
      const addIframeCreationSniffer = () => {
        const createElementHandler = {
          // Make toString() native
          get(target, key) {
            return Reflect.get(target, key);
          },
          apply(target, thisArg, args) {
            if (`${args[0]}`.toLowerCase() === 'iframe') {
              // Everything as usual
              return handleIframeCreation(target, thisArg, args);
            }
            return target.apply(thisArg, args);
          },
        };
        // All this just due to iframes with srcdoc bug
        overridePropertyWithProxy(document, 'createElement', createElementHandler);
      };
      // Let's go
      addIframeCreationSniffer();
    } catch (err) {
      // warning message supressed (see https://github.com/apify/fingerprint-suite/issues/61).
      // console.warn(err)
    }
  }

  function fixPluginArray() {
    if (window.navigator.plugins.length !== 0) {
      return;
    }
    Object.defineProperty(navigator, 'plugins', {
      get: () => {
        const ChromiumPDFPlugin = Object.create(Plugin.prototype, {
          description: {
            value: 'Portable Document Format',
            enumerable: false,
          },
          filename: {
            value: 'internal-pdf-viewer',
            enumerable: false
          },
          name: {
            value: 'Chromium PDF Plugin',
            enumerable: false
          },
        });
        return Object.create(PluginArray.prototype, {
          length: {
            value: 1
          },
          0: {
            value: ChromiumPDFPlugin
          },
        });
      },
    });
  }

  function runHeadlessFixes() {
    try {
      if (isHeadlessChromium) {
        fixWindowChrome();
        fixPermissions();
        fixIframeContentWindow();
        fixPluginArray();
      }
    } catch (e) {
      console.error(e);
    }
  }

  function overrideStatic() {
    try {
      window.SharedArrayBuffer = undefined;
    } catch (e) {
      console.error(e);
    }
  }
  //# sourceMappingURL=utils.js.map
  ;
  const fp = {
    "screen": {
      "availTop": 25,
      "availLeft": 0,
      "pageXOffset": 0,
      "pageYOffset": 0,
      "screenX": 32,
      "hasHDR": false,
      "width": 1680,
      "height": 1050,
      "availWidth": 1680,
      "availHeight": 1025,
      "clientWidth": 0,
      "clientHeight": 19,
      "innerWidth": 0,
      "innerHeight": 0,
      "outerWidth": 1626,
      "outerHeight": 968,
      "colorDepth": 24,
      "pixelDepth": 24,
      "devicePixelRatio": 2
    },
    "audioCodecs": {
      "ogg": "probably",
      "mp3": "probably",
      "wav": "probably",
      "m4a": "maybe",
      "aac": "probably"
    },
    "videoCodecs": {
      "ogg": "",
      "h264": "probably",
      "webm": "probably"
    },
    "pluginsData": {
      "plugins": [{
        "name": "PDF Viewer",
        "description": "Portable Document Format",
        "filename": "internal-pdf-viewer",
        "mimeTypes": [{
          "type": "application/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "PDF Viewer"
        }, {
          "type": "text/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "PDF Viewer"
        }]
      }, {
        "name": "Chrome PDF Viewer",
        "description": "Portable Document Format",
        "filename": "internal-pdf-viewer",
        "mimeTypes": [{
          "type": "application/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Chrome PDF Viewer"
        }, {
          "type": "text/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Chrome PDF Viewer"
        }]
      }, {
        "name": "Chromium PDF Viewer",
        "description": "Portable Document Format",
        "filename": "internal-pdf-viewer",
        "mimeTypes": [{
          "type": "application/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Chromium PDF Viewer"
        }, {
          "type": "text/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Chromium PDF Viewer"
        }]
      }, {
        "name": "Microsoft Edge PDF Viewer",
        "description": "Portable Document Format",
        "filename": "internal-pdf-viewer",
        "mimeTypes": [{
          "type": "application/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Microsoft Edge PDF Viewer"
        }, {
          "type": "text/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "Microsoft Edge PDF Viewer"
        }]
      }, {
        "name": "WebKit built-in PDF",
        "description": "Portable Document Format",
        "filename": "internal-pdf-viewer",
        "mimeTypes": [{
          "type": "application/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "WebKit built-in PDF"
        }, {
          "type": "text/pdf",
          "suffixes": "pdf",
          "description": "Portable Document Format",
          "enabledPlugin": "WebKit built-in PDF"
        }]
      }],
      "mimeTypes": ["Portable Document Format~~application/pdf~~pdf", "Portable Document Format~~text/pdf~~pdf"]
    },
    "battery": {
      "charging": true,
      "chargingTime": 0,
      "dischargingTime": null,
      "level": 1
    },
    "videoCard": {
      "renderer": "ANGLE (Intel, ANGLE Metal Renderer: Intel(R) Iris(TM) Plus Graphics 655, Unspecified Version)",
      "vendor": "Google Inc. (Intel)"
    },
    "multimediaDevices": {
      "speakers": [{
        "deviceId": "",
        "kind": "audiooutput",
        "label": "",
        "groupId": ""
      }],
      "micros": [{
        "deviceId": "",
        "kind": "audioinput",
        "label": "",
        "groupId": ""
      }],
      "webcams": [{
        "deviceId": "",
        "kind": "videoinput",
        "label": "",
        "groupId": ""
      }]
    },
    "fonts": ["Arial Unicode MS", "Gill Sans", "Helvetica Neue", "Menlo"],
    "mockWebRTC": false,
    "slim": false,
    "navigator": {
      "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
      "userAgentData": {
        "brands": [{
          "brand": "Not;A=Brand",
          "version": "99"
        }, {
          "brand": "Google Chrome",
          "version": "139"
        }, {
          "brand": "Chromium",
          "version": "139"
        }],
        "mobile": false,
        "platform": "macOS"
      },
      "language": "en-US",
      "languages": ["en-US", "en"],
      "platform": "MacIntel",
      "deviceMemory": 8,
      "hardwareConcurrency": 10,
      "maxTouchPoints": 0,
      "product": "Gecko",
      "productSub": "20030107",
      "vendor": "Google Inc.",
      "vendorSub": null,
      "doNotTrack": null,
      "appCodeName": "Mozilla",
      "appName": "Netscape",
      "appVersion": "5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
      "oscpu": null,
      "extraProperties": {
        "vendorFlavors": ["chrome"],
        "globalPrivacyControl": null,
        "pdfViewerEnabled": true,
        "installedApps": []
      },
      "webdriver": false
    },
    "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
    "historyLength": 2
  };
  (function inject() {
    const {
      battery,
      navigator: {
        extraProperties,
        userAgentData,
        webdriver,
        ...navigatorProps
      },
      screen: allScreenProps,
      videoCard,
      historyLength,
      audioCodecs,
      videoCodecs,
      mockWebRTC,
      slim,
      // @ts-expect-error internal browser code
    } = fp;
    const {
      // window screen props
      outerHeight,
      outerWidth,
      devicePixelRatio,
      innerWidth,
      innerHeight,
      screenX,
      pageXOffset,
      pageYOffset,
      // Document screen props
      clientWidth,
      clientHeight,
      // Ignore hdr for now.
      hasHDR,
      // window.screen props
      ...newScreen
    } = allScreenProps;
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
    if (mockWebRTC)
      blockWebRTC();
    if (slim) {
      // @ts-expect-error internal browser code
      // eslint-disable-next-line dot-notation
      window['slim'] = true;
    }
    overrideIntlAPI(navigatorProps.language);
    overrideStatic();
    if (userAgentData) {
      // overrideUserAgentData(userAgentData);
    }
    if (window.navigator.webdriver) {
      navigatorProps.webdriver = false;
    }
    overrideInstancePrototype(window.navigator, {
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
    });
    overrideInstancePrototype(window.screen, newScreen);
    overrideWindowDimensionsProps(windowScreenProps);
    overrideDocumentDimensionsProps(documentScreenProps);
    overrideInstancePrototype(window.history, {
      length: historyLength,
    });
    // overrideWebGl(videoCard);
    overrideCodecs(audioCodecs, videoCodecs);
    overrideBattery(battery);
  })();
  (function inject2() {
    var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
      if (pack || arguments.length === 2)
        for (var i = 0, l = from.length, ar; i < l; i++) {
          if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
          }
        }
      return to.concat(ar || Array.prototype.slice.call(from));
    };

    var seededRandom = function (seed, max, min) {
      if (max === void 0) {
        max = 1;
      }
      if (min === void 0) {
        min = 0;
      }
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
      var m = _array.length,
        t, i;
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
      return arr.filter(function (item) {
        return item !== undefined && item !== false;
      });
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
          const noiseStrength = 1 + (seed % 3); // 1～3 强度

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
          if (prop === symbolRaw) return t; // 外部可拿到原对象
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
      overrideGetterWithProxy(DOMRectReadOnly.prototype, "top", getHandler((rect) => {
        return rect.y;
      }))
      overrideGetterWithProxy(DOMRectReadOnly.prototype, "left", getHandler((rect) => {
        return rect.x;
      }))
      overrideGetterWithProxy(DOMRectReadOnly.prototype, "bottom", getHandler((rect) => {
        return rect.y + rect.height;
      }))
      overrideGetterWithProxy(DOMRectReadOnly.prototype, "right", getHandler((rect) => {
        return rect.x + rect.width;
      }))

      overridePropertyWithProxy(DOMRectReadOnly.prototype, 'toJSON', {
        apply: function (target, thisArg, args) {
          return pick(thisArg, ['x', 'y', 'width', 'height', 'bottom', 'left', 'right', 'top']);
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
  })();
})();