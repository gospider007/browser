enum HookType {
  default = 0,  // 系统值
  value = 1,  // 自定义值
  page = 2,  // 每个标签页随机
  browser = 3,  // 每次启动浏览器随机
  domain = 4,  // 根据域名随机
  global = 5,  // 根据全局种子随机
  enabled = 6,  // 启用
  disabled = 7,  // 禁用
}
/**
 * 线性同余，根据seed产生随机数
 */
const seededRandom = function (seed: number | string, max: number = 1, min: number = 0): number {
  if (typeof seed === 'string') {
    seed = hashNumberFromString(seed);
  }
  const mod = 233280;
  seed = (seed * 9301 + 49297) % mod;
  if (seed < 0) seed += mod; // 确保 seed 为正数
  const rnd = seed / mod;
  return min + rnd * (max - min);
}

/**
 * 根据种子随机获取数组中的元素
 */
const seededEl = <T>(arr: Readonly<T[]>, seed: number): T => {
  return arr[seed % arr.length];
}

/**
 * 数组洗牌
 */
const shuffleArray = <T>(array: Readonly<T[]>, seed: number): T[] => {
  const _array = [...array];
  let m = _array.length, t: T, i: number;

  const random = () => {
    const x = Math.sin(seed++) * 10000;
    return x - Math.floor(x);
  };

  while (m) {
    i = Math.floor(random() * m--);
    t = _array[m];
    _array[m] = _array[i];
    _array[i] = t;
  }

  return _array;
}

/**
 * 版本号比较
 * @returns v1大于v2，返回1；v1小于v2，返回-1；v1等于v2，返回0
 */
const compareVersions = function (v1: string, v2: string): -1 | 0 | 1 {
  const parse = (v: string) => {
    const [main, pre = ''] = v.split('-');
    const parts = main.split('.').map(n => parseInt(n, 10));
    return { parts, pre };
  };

  const { parts: p1, pre: pre1 } = parse(v1);
  const { parts: p2, pre: pre2 } = parse(v2);
  const maxLen = Math.max(p1.length, p2.length);

  for (let i = 0; i < maxLen; i++) {
    const a = p1[i] ?? 0;
    const b = p2[i] ?? 0;
    if (a > b) return 1;
    if (a < b) return -1;
  }

  if (pre1 && !pre2) return -1; // 预发布版本 < 正式版本
  if (!pre1 && pre2) return 1;
  if (pre1 && pre2) return pre1.localeCompare(pre2) as -1 | 0 | 1;

  return 0;
}

/**
 * 生成随机的种子
 */
const genRandomSeed = function () {
  return Math.floor(seededRandom(Math.random() * 1e6, Number.MAX_SAFE_INTEGER, 1))
}

/**
 * 字符串的number类型hash
 */
const hashNumberFromString = (input: string): number => {
  let hash = 0;
  for (let i = 0; i < input.length; i++) {
    let char = input.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return Math.abs(hash % Number.MAX_SAFE_INTEGER);
}

/**
 * 过滤arr中的undefined和false值
 */
const arrayFilter = function <T>(arr: (T | undefined | boolean)[]): T[] {
  return arr.filter(item => item !== undefined && item !== false) as T[]
}

const tryUrl = (url: string) => {
  try {
    return new URL(url);
  } catch (err) {
    return undefined
  }
}

/**
 * 版本号随机偏移
 * @param sourceVersion 源版本号
 * @param seed 种子
 * @param maxSubVersionNumber 最大子版本数量 
 * @param mainVersionOffset 最大主版本号偏移
 * @param subVersionOffset 最大子版本号偏移
 * @returns 
 */
const versionRandomOffset = (sourceVersion: string, seed: number, maxSubVersionNumber?: number, maxMainVersionOffset?: number, maxSubVersionOffset?: number): string => {
  // 将源版本号分解为主版本号和子版本号
  const [mainVersion, ...subversions] = sourceVersion.split('.')
  if (mainVersion === undefined) return sourceVersion
  let nMainVersion = Number(mainVersion)
  if (Number.isNaN(nMainVersion)) return sourceVersion

  maxMainVersionOffset = maxMainVersionOffset ?? 2
  maxSubVersionOffset = maxSubVersionOffset ?? 50
  maxSubVersionNumber = maxSubVersionNumber ?? subversions.length

  nMainVersion += (seed % ((maxMainVersionOffset * 2) + 1)) - maxMainVersionOffset;

  const nSubversions: string[] = []
  for (let i = 0; i < maxSubVersionNumber; i++) {
    const subversion = subversions[i]
    let nSubversion = Number(subversion)
    if (Number.isNaN(nSubversion)) {
      nSubversions.push(subversion)
      continue
    }
    const ss = Math.floor(seededRandom(seed + i, -maxSubVersionOffset, maxSubVersionOffset))
    nSubversion = Math.abs((nSubversion ?? 0) + ss)
    nSubversions.push(nSubversion.toString())
  }

  // 将主版本号和子版本号重新组合成完整的版本号
  return [nMainVersion, ...nSubversions].join('.');
}

/**
 * 获取主要版本号
 */
const getMainVersion = (sourceVersion: string) => {
  return sourceVersion.split('.')[0]
}

/**
 * 向上取10的幂次
 */
const getNextPowerOfTen = (num: number) => {
  if (num === 0) return 0;
  else if (num === 1) return 10;

  if (num < 0) num = -num;
  return Math.pow(10, Math.ceil(Math.log10(num)));
}


/**
 * 深度代理
 */
const deepProxy = <T>(obj: T, handler: ProxyHandler<any>, seen = new WeakMap<any>()): T => {
  if (seen.has(obj)) { return seen.get(obj) }
  const proxy = new Proxy(obj, {
    get(target, property, receiver) {
      const value = target[property];
      if (typeof value === 'object' && value !== null) {
        return deepProxy(value, handler);
      }
      return handler.get ? handler.get(target, property, receiver) : value;
    },
    set: handler.set,
  });
  seen.set(obj, proxy);
  return proxy;
}

/**
 * 是否存在父域名或自身
 * @param src 子域名 
 */
const existParentDomain = (domains: string[], src: string) => {
  if (!src) return false;
  if (!domains?.length) return false;
  src = '.' + src
  return domains.some((v) => src.endsWith('.' + v))
}

/**
 * 是否存在子域名或自身
 * @param src 父域名
 */
const existChildDomain = (domains: string[], src: string) => {
  if (!src) return false;
  if (!domains?.length) return false;
  src = '.' + src
  return domains.some((v) => ('.' + v).endsWith(src))
}

/**
 * 查找子域名和自身
 * @param src 父域名
 */
const selectChildDomains = (domains: string[], src: string) => {
  if (!src) return []
  if (!domains?.length) return []
  src = '.' + src
  const list: string[] = []
  for (const domain of domains) {
    if (('.' + domain).endsWith(src)) list.push(domain);
  }
  return list
}

/**
 * 查找父域名和自身
 * @param src 子域名 
 */
const selectParentDomains = (domains: string[], src: string) => {
  if (!src) return []
  if (!domains?.length) return []
  src = '.' + src
  const list: string[] = []
  for (const domain of domains) {
    if (src.endsWith('.' + domain)) list.push(domain);
  }
  return list
}

/**
 * 对象提取值
 */
const pick = <T extends object, K extends keyof T>(obj: T, keys: readonly K[]): Pick<T, K> => {
  const res = {} as Pick<T, K>;
  for (const key of keys) {
    res[key] = obj[key];
  }
  return res;
}

/**
 * 移除后缀
 */
function trimSuffix(str: string) {
  const index = str.lastIndexOf('.');
  return index !== -1 ? str.substring(0, index) : str;
}




// 
// --- notification ---
// 
let fpNoticePool: Record<string, number> = {}
let iframeNoticePool: Record<string, number> = {}

/**
 * 记录指纹数量
 */
const notify = (key: string) => {
  fpNoticePool[key] = (fpNoticePool[key] ?? 0) + 1
}



/**
 * 记录iframe数量
 */
const notifyIframeOrigin = (key?: string) => {
  if (!key || key === 'null') key = 'about:blank';
  iframeNoticePool[key] = (iframeNoticePool[key] ?? 0) + 1
}


// 
// --- random ---
// 

/**
 * 随机canvas噪音
 */
const randomCanvasNoise = (seed: number) => {
  const noise: number[] = []
  for (let i = 0; i < 10; i++) {
    noise.push(Math.floor(seededRandom(seed++, 255, 0)))
  }
  return noise
}

/**
 * 获取[x, y]，区间[-1, 1]
 */
const randomWebglNoise = (seed: number): [number, number] => {
  return [seededRandom(seed, 1, -1), seededRandom(seed + 1, 1, -1)]
}

/**
 * 获取随机字体噪音
 */
const randomFontNoise = (seed: number, mark: string): number => {
  const random = seededRandom((seed + hashNumberFromString(mark)) % Number.MAX_SAFE_INTEGER, 3, 0)
  if ((random * 10) % 1 < 0.9) return 0;
  return Math.floor(random) - 1;
}

/**
 * 获取随机屏幕尺寸
 */
const randomScreenSize = (screen: Screen, seed: number) => {
  const offset = Math.round(seededRandom(seed, 100, -100))
  const ratio = screen.width / screen.height;

  let width: number
  let height: number
  if (screen.width >= screen.height) {
    // 偏移宽度
    width = screen.width + offset;
    height = Math.round(width / ratio);
  } else {
    // 偏移高度
    height = screen.height + offset;
    width = Math.round(height * ratio);
  }

  return { width, height };
}


// 
// --- other ---
// 

type U8Array = Uint8ClampedArray | Uint8Array;

const isPixelEqual = (p1: U8Array, p2: U8Array) => {
  return p1[0] === p2[0] && p1[1] === p2[1] && p1[2] === p2[2] && p1[3] === p2[3];
}

const pixelCopy = (src: U8Array, dst: U8Array, index: number) => {
  dst[0] = src[index]
  dst[1] = src[index + 1]
  dst[2] = src[index + 2]
  dst[3] = src[index + 3]
}

/**
 * 在2d画布绘制噪音
 */
const drawNoise = (
  rawFunc: typeof CanvasRenderingContext2D.prototype.getImageData,
  noise: number[],
  ctx: CanvasRenderingContext2D,
  sx: number, sy: number, sw: number, sh: number, settings?: ImageDataSettings
) => {
  const imageData = rawFunc.call(ctx, sx, sy, sw, sh, settings)

  let noiseIndex = 0;
  let isChanged = false

  const Arr = Uint8ClampedArray;
  const center = new Arr(4)
  const up = new Arr(4)
  const down = new Arr(4)
  const left = new Arr(4)
  const right = new Arr(4)

  const pixelData = imageData.data

  outer: for (let row = 1; row < sh - 2; row += 2) {
    for (let col = 1; col < sw - 2; col += 2) {
      if (noise.length === noiseIndex) { break outer; }

      const index = (row * sw + col) * 4;
      pixelCopy(pixelData, center, index)

      pixelCopy(pixelData, up, ((row - 1) * sw + col) * 4)
      if (isPixelEqual(center, up)) continue;

      pixelCopy(pixelData, down, ((row + 1) * sw + col) * 4)
      if (isPixelEqual(center, down)) continue;

      pixelCopy(pixelData, left, (row * sw + (col - 1)) * 4)
      if (isPixelEqual(center, left)) continue;

      pixelCopy(pixelData, right, (row * sw + (col + 1)) * 4)
      if (isPixelEqual(center, right)) continue;

      pixelData[index + 3] = (noise[noiseIndex++] % 256)
      isChanged = true
    }
  }

  if (isChanged) {
    ctx.putImageData(imageData, sx, sy)
  }

  return imageData
}


/**
 * 在webgl上下文绘制噪音点
 * @param noisePosition 区间[-1, 1]
 */
const drawNoiseToWebgl = (gl: WebGLRenderingContext | WebGL2RenderingContext, noisePosition: [number, number]) => {
  const vertexShaderSource = `attribute vec4 noise;void main() {gl_Position = noise;gl_PointSize = 0.001;}`;
  const fragmentShaderSource = `void main() {gl_FragColor = vec4(0.0, 0.0, 0.0, 0.01);}`;

  const createShader = (gl: WebGLRenderingContext | WebGL2RenderingContext, type: GLenum, source: string) => {
    const shader = gl.createShader(type);
    if (!shader) return;
    gl.shaderSource(shader, source);
    gl.compileShader(shader);
    return shader;
  }

  const vertexShader = createShader(gl, gl.VERTEX_SHADER, vertexShaderSource);
  const fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, fragmentShaderSource);
  if (!vertexShader || !fragmentShader) return;

  const program = gl.createProgram();
  if (!program) return;
  gl.attachShader(program, vertexShader);
  gl.attachShader(program, fragmentShader);
  gl.linkProgram(program);
  gl.useProgram(program);

  const positions = new Float32Array(noisePosition);
  const positionBuffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
  gl.bufferData(gl.ARRAY_BUFFER, positions, gl.STATIC_DRAW);

  const noise = gl.getAttribLocation(program, 'noise');
  gl.enableVertexAttribArray(noise);
  gl.vertexAttribPointer(noise, 2, gl.FLOAT, false, 0, 0);
  gl.drawArrays(gl.POINTS, 0, 1);
}


const hookTasks  = [
  /**
   * Canvas
   */
  {
    onEnable: ({ win, conf, useSeed, useProxy }) => {
      /* getContext */
      useProxy(win.HTMLCanvasElement.prototype, 'getContext', {
        apply: (target, thisArg, args: Parameters<typeof HTMLCanvasElement.prototype.getContext>) => {
          if (args[0] === '2d') {
            const option = args[1] ?? {};
            option.willReadFrequently = true;
            args[1] = option
          }
          return target.apply(thisArg, args);
        }
      })

      /* getImageData */
      {
        const seed = useSeed(conf.fp.other.canvas)
        if (seed != null) {
          const noise = randomCanvasNoise(seed)
          useProxy(win.CanvasRenderingContext2D.prototype, 'getImageData', {
            apply: (target, thisArg: CanvasRenderingContext2D, args: Parameters<typeof CanvasRenderingContext2D.prototype.getImageData>) => {
              notify('strong.canvas')
              return drawNoise(target, noise, thisArg, ...args);
            }
          })
        }
      }
    },
  },

  /**
   * Webgl
   */
  {
    condition: ({ conf }) => conf.fp.other.webgl.type !== HookType.default,
    onEnable: ({ win, conf, useSeed, useProxy }) => {
      /* Image */
      {
        const seed = useSeed(conf.fp.other.webgl)
        if (seed != null) {
          const noise = randomWebglNoise(seed)
          const handler = {
            apply: (target: any, thisArg: WebGLRenderingContext | WebGL2RenderingContext, args: any) => {
              notify('strong.webgl')
              drawNoiseToWebgl(thisArg, noise)
              return target.apply(thisArg, args as any);
            }
          }
          useProxy(win.WebGLRenderingContext.prototype, 'readPixels', handler)
          useProxy(win.WebGL2RenderingContext.prototype, 'readPixels', handler)
        }
      }
      /* Report: Supported Extensions */
      {
        const seed = useSeed(conf.fp.other.webgl)
        if (seed != null) {
          const noise = seededRandom(seed, 1, 0)
          const handler = {
            apply: (target: any, thisArg: WebGLRenderingContext, args: any) => {
              notify('strong.webgl')
              const res = target.apply(thisArg, args);
              res?.push?.('EXT_' + noise);
              return res;
            }
          }
          useProxy(win.WebGLRenderingContext.prototype, 'getSupportedExtensions', handler)
          useProxy(win.WebGL2RenderingContext.prototype, 'getSupportedExtensions', handler)
        }
      }
    },
  },

  /**
   * Webgl参数信息
   */
  {
    condition: ({ conf, isDefault }) => !isDefault(conf.fp.normal.gpuInfo),
    onEnable: ({ win, conf, useHookMode, useProxy }) => {
      const fps = conf.fp.normal

      let ex: WEBGL_debug_renderer_info | null

      /* Report: Parameter */
      const info = useHookMode(fps.gpuInfo).value
      if (info) {
        const handler = {
          apply: (target: any, thisArg: WebGLRenderingContext, args: any) => {
            if (!ex) ex = thisArg.getExtension('WEBGL_debug_renderer_info');
            if (ex) {
              if (args[0] === ex.UNMASKED_VENDOR_WEBGL) {
                notify('weak.gpuInfo')
                // 模拟调用
                if (info.vendor && target.apply(thisArg, args)) {
                  return info.vendor;
                }
              } else if (args[0] === ex.UNMASKED_RENDERER_WEBGL) {
                notify('weak.gpuInfo')
                if (info.renderer && target.apply(thisArg, args)) {
                  return info.renderer;
                }
              }
            }
            return target.apply(thisArg, args);
          }
        }
        useProxy(win.WebGLRenderingContext.prototype, 'getParameter', handler)
        useProxy(win.WebGL2RenderingContext.prototype, 'getParameter', handler)
      }
    }
  },

  /**
   * toDataURL
   */
  {
    condition: ({ conf, isDefault }) => !isDefault([conf.fp.other.canvas, conf.fp.other.webgl]),
    onEnable: ({ win, conf, useSeed, useProxy, useRaw }) => {

      const seedCanvas = useSeed(conf.fp.other.canvas)
      const seedWebgl = useSeed(conf.fp.other.webgl)
      const noiseCanvas = seedCanvas == null ? null : randomCanvasNoise(seedCanvas)
      const noiseWebgl = seedWebgl == null ? null : randomWebglNoise(seedWebgl)

      useProxy(win.HTMLCanvasElement.prototype, 'toDataURL', {
        apply: (target, thisArg: HTMLCanvasElement, args: Parameters<typeof HTMLCanvasElement.prototype.toDataURL>) => {
          /* 2d */
          if (noiseCanvas) {
            const ctx = thisArg.getContext('2d');
            if (ctx) {
              notify('strong.canvas')
              drawNoise(
                useRaw(win.CanvasRenderingContext2D.prototype.getImageData),
                noiseCanvas, ctx,
                0, 0, thisArg.width, thisArg.height
              )
              return target.apply(thisArg, args);
            }
          }
          /* webgl */
          if (noiseWebgl) {
            const gl = thisArg.getContext('webgl') ?? thisArg.getContext('webgl2')
            if (gl) {
              notify('strong.webgl')
              noiseWebgl && drawNoiseToWebgl(gl as any, noiseWebgl)
              return target.apply(thisArg, args);
            }
          }
          return target.apply(thisArg, args);
        }
      })

    },
  },

  /**
   * Audio
   * 音频指纹
   */
  {
    condition: ({ conf }) => conf.fp.other.audio.type !== HookType.default,
    onEnable: ({ win, conf, useSeed, useProxy }) => {
      const seed = useSeed(conf.fp.other.audio)
      if (seed == null) return;

      const noise = seededRandom(seed, 1, 0)
      useProxy(win.OfflineAudioContext.prototype, 'createDynamicsCompressor', {
        apply: (target, thisArg: OfflineAudioContext, args: Parameters<typeof OfflineAudioContext.prototype.createDynamicsCompressor>) => {
          notify('strong.audio')
          const compressor = target.apply(thisArg, args)
          const gain = thisArg.createGain()
          gain.gain.value = noise * 0.001
          compressor.connect(gain)
          gain.connect(thisArg.destination)
          return compressor
        }
      });

    },
  },

  /**
   * Timezone
   * 时区
   */
  {
    condition: ({ conf }) => conf.fp.other.timezone.type !== HookType.default,
    onEnable: ({ win, conf, useHookMode, useProxy }) => {
      const tzValue = useHookMode(conf.fp.other.timezone).value
      if (!tzValue) return;

      const _DateTimeFormat = win.Intl.DateTimeFormat;

      type TimeParts = Partial<Record<keyof Intl.DateTimeFormatPartTypesRegistry, string>>
      const getStandardDateTimeParts = (date: Date): TimeParts | null => {
        const formatter = new _DateTimeFormat('en-US', {
          timeZone: tzValue.zone ?? 'Asia/Shanghai',
          weekday: 'short',
          month: 'short',
          day: '2-digit',
          year: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          fractionalSecondDigits: 3,
          hour12: false,
          timeZoneName: 'longOffset',
        })
        try {
          const parst = formatter.formatToParts(date)
          return parst.reduce((acc: TimeParts, cur) => {
            acc[cur.type] = cur.value
            return acc
          }, {})
        } catch (e) {
          return null
        }
      }

      /* DateTimeFormat */
      useProxy(win.Intl, 'DateTimeFormat', {
        construct: (target, args: Parameters<typeof Intl.DateTimeFormat>, newTarget) => {
          notify('weak.timezone')
          args[0] = args[0] ?? tzValue.locale
          args[1] = Object.assign({ timeZone: tzValue.zone }, args[1]);
          return new target(...args)
        },
        apply: (target, thisArg: Intl.DateTimeFormat, args: Parameters<typeof Intl.DateTimeFormat>) => {
          notify('weak.timezone')
          args[0] = args[0] ?? tzValue.locale
          args[1] = Object.assign({ timeZone: tzValue.zone }, args[1]);
          return target.apply(thisArg, args)
        },
      })


      /* Date */
      useProxy(win, 'Date', {
        apply: (target, thisArg: Date, args: Parameters<typeof Date>) => {
          return new target(...args).toString()
        }
      })

      /* getTimezoneOffset & toString */
      {
        const tasks: { [key in keyof Date]?: (thisArg: Date) => any } = {
          'getTimezoneOffset': (_) => tzValue.offset * -60,
          'toString': (thisArg) => {
            const ps = getStandardDateTimeParts(thisArg)
            return ps && `${ps.weekday} ${ps.month} ${ps.day} ${ps.year} ${ps.hour}:${ps.minute}:${ps.second} ${ps.timeZoneName?.replace(':', '')}`
          },
          'toDateString': (thisArg) => {
            const ps = getStandardDateTimeParts(thisArg)
            return ps && `${ps.weekday} ${ps.month} ${ps.day} ${ps.year}`
          },
          'toTimeString': (thisArg) => {
            const ps = getStandardDateTimeParts(thisArg)
            return ps && `${ps.hour}:${ps.minute}:${ps.second} ${ps.timeZoneName?.replace(':', '')}`
          },
        }
        useProxy(win.Date.prototype,
          Object.keys(tasks) as (keyof Date)[],
          (key) => {
            const task = tasks[key]
            return task && {
              apply: (target: any, thisArg: Date, args: Parameters<typeof Date.prototype.toString>) => {
                notify('weak.timezone')
                const result = task(thisArg)
                return result == null ? target.apply(thisArg, args) : result
              }
            }
          })
      }

      /* toLocaleString */
      useProxy(win.Date.prototype, [
        'toLocaleString', 'toLocaleDateString', 'toLocaleTimeString'
      ], {
        apply: (target: any, thisArg: Date, args: Parameters<typeof Date.prototype.toLocaleString>) => {
          notify('weak.timezone')
          args[0] = args[0] ?? tzValue.locale
          args[1] = Object.assign({ timeZone: tzValue.zone }, args[1]);
          return target.apply(thisArg, args);
        }
      })

    },
  },

  /**
   * Webrtc
   */
  {
    condition: ({ conf }) => conf.fp.other.webrtc.type !== HookType.default,
    onEnable: ({ win, useDefine }) => {

      useDefine([win.Navigator.prototype, win.navigator], 'mediaDevices', {
        get() { return null }
      });

      [
        'getUserMedia',
        'mozGetUserMedia',
        'webkitGetUserMedia'
      ].forEach((key) => {
        // @ts-ignore
        if (win.Navigator.prototype[key]) win.Navigator.prototype[key] = undefined;
      });

      [
        'RTCDataChannel',
        'RTCIceCandidate',
        'RTCConfiguration',
        'MediaStreamTrack',
        'RTCPeerConnection',
        'RTCSessionDescription',
        'mozMediaStreamTrack',
        'mozRTCPeerConnection',
        'mozRTCSessionDescription',
        'webkitMediaStreamTrack',
        'webkitRTCPeerConnection',
        'webkitRTCSessionDescription',
      ].forEach((key) => {
        // @ts-ignore
        if (win[key]) win[key] = undefined;
      });
    },
  },

  /**
   * Font
   * 字体指纹
   */
  {
    condition: ({ conf }) => conf.fp.other.font.type !== HookType.default,
    onEnable: ({ win, conf, useSeed, useProxy, useGetterProxy }) => {
      const seed = useSeed(conf.fp.other.font)
      if (seed == null) return;

      useGetterProxy(win.HTMLElement.prototype, [
        'offsetHeight', 'offsetWidth'
      ], (key, getter) => ({
        apply(target: () => any, thisArg: HTMLElement, args: any) {
          notify('strong.fonts')
          const result = getter.call(thisArg);
          const mark = (thisArg.style?.fontFamily ?? key) + result;
          return result + randomFontNoise(seed, mark);
        }
      }))

      useProxy(win, 'FontFace', {
        construct: (target, args: ConstructorParameters<typeof FontFace>, newTarget) => {
          const source = args[1]
          if (typeof source === 'string' && source.startsWith('local(')) {
            notify('strong.fonts')
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
      })

    },
  },

  /**
   * Webgpu
   */
  {
    condition: ({ conf }) => conf.fp.other.webgpu.type !== HookType.default,
    onEnable: ({ win, conf, useSeed, useDefine, useProxy, newProxy }) => {
      const seed = useSeed(conf.fp.other.webgpu)
      if (seed == null) return;

      /* GPUAdapter & GPUDevice */
      {
        const makeNoise = (raw: any, offset: number) => {
          notify('strong.webgpu')
          const rn = seededRandom(seed + (offset * 7), 64, 1)
          return raw ? raw - Math.floor(rn) : raw;
        }

        const handler = (_: any, desc: PropertyDescriptor) => {
          const getter = desc.get
          return getter && {
            get() {
              const limits = getter.call(this);
              return newProxy(limits, {
                get(target, prop) {
                  const value = target[prop];
                  switch (prop) {
                    case "maxBufferSize": return makeNoise(value, 0);
                    case "maxStorageBufferBindingSize": return makeNoise(value, 1);
                  }
                  return typeof value === "function" ? value.bind(target) : value;
                }
              })
            }
          }
        }

        // @ts-ignore
        win.GPUAdapter && useDefine(win.GPUAdapter.prototype, 'limits', handler)
        // @ts-ignore
        win.GPUDevice && useDefine(win.GPUDevice.prototype, 'limits', handler)
      }

      /*** GPUCommandEncoder ***/
      // @ts-ignore
      if (win.GPUCommandEncoder?.prototype?.beginRenderPass) {
        // @ts-ignore
        useProxy(win.GPUCommandEncoder.prototype, 'beginRenderPass', {
          apply(target, self, args) {
            notify('strong.webgpu')
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

      /*** GPUQueue ***/
      // @ts-ignore
      if (win.GPUQueue?.prototype?.writeBuffer) {
        // @ts-ignore
        useProxy(win.GPUQueue.prototype, 'writeBuffer', {
          apply(target, self, args) {
            notify('strong.webgpu')
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
    },
  },

  /**
   * DomRect
   */
  {
    onEnable: ({ win, conf, useSeed, useProxy, useGetterProxy }) => {
      const seed = useSeed(conf.fp.other.domRect)
      if (seed == null) return;

      const mine = new WeakSet<DOMRect>()

      useProxy(win, [
        'DOMRect', 'DOMRectReadOnly'
      ], {
        construct(target, args, newTarget) {
          const res = Reflect.construct(target, args, newTarget)
          // mine.add(res)
          return res;
        }
      })

      {
        const noise = seededRandom(seed, 1e-6, -1e-6);
        useGetterProxy(win.DOMRect.prototype, [
          'x', 'y', 'width', 'height'
        ], (_, getter) => ({
          apply(target, thisArg: DOMRect, args: any) {
            notify('strong.domRect')
            const res = getter.call(thisArg);
            if (mine.has(thisArg)) return res;
            return res + noise;
          }
        }))
      }

      {
        const hook = (key: keyof DOMRectReadOnly, toResult: (rect: DOMRectReadOnly) => number) => {
          useGetterProxy(win.DOMRectReadOnly.prototype, key, () => ({
            apply(target, thisArg: DOMRectReadOnly, args: any) {
              return toResult(thisArg)
            }
          }))
        }
        hook('top', rect => rect.y)
        hook('left', rect => rect.x)
        hook('bottom', rect => rect.y + rect.height)
        hook('right', rect => rect.x + rect.width)
      }

      useProxy(win.DOMRectReadOnly.prototype, 'toJSON', {
        apply(target, thisArg: DOMRectReadOnly, args: any) {
          notify('strong.domRect')
          return pick(thisArg, ['x', 'y', 'width', 'height', 'bottom', 'left', 'right', 'top']);
        }
      })
    }
  },
];





















({ win, conf, useSeed, useProxy }) => {
      if (!win) return;

      const seed = useSeed(conf.fp.other.domRect)
      if (seed == null) return;

      const noise = seededRandom(seed, 1e-6, -1e-6);

      {
        const handler = {
          apply(target: () => DOMRect, thisArg: any, args: any) {
            notify('strong.domRect')
            const rect = Reflect.apply(target, thisArg, args);
            if (rect) {
              if (rect.x !== 0) rect.x += noise;
              if (rect.width !== 0) rect.width += noise;
            }
            return rect;
          }
        }
        useProxy(win.Element.prototype, 'getBoundingClientRect', handler)
        useProxy(win.Range.prototype, 'getBoundingClientRect', handler)
      }

      {
        const handler = {
          apply(target: () => DOMRectList, thisArg: any, args: any) {
            notify('strong.domRect')
            const rlist = Reflect.apply(target, thisArg, args);
            if (rlist) {
              for (let i = 0; i < rlist.length; i++) {
                const rect = rlist[i];
                if (rect.x !== 0) rect.x += noise;
                if (rect.width !== 0) rect.width += noise;
              }
            }
            return rlist;
          }
        }
        useProxy(win.Element.prototype, 'getClientRects', handler)
        useProxy(win.Range.prototype, 'getClientRects', handler)
      }
    }
