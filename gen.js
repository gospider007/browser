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














/**
 * 可复现随机数
 */
function createPRNG(seed) {
    let x = seed | 0;
    return function () {
        x ^= x << 13;
        x ^= x >> 17;
        x ^= x << 5;
        return (x >>> 0) / 4294967296;
    };
}

function drawNoise(target, noise, ctx, x, y, w, h) {
    const prng = createPRNG(noise);  // 可复现随机
    // 原始图像
    const img = target.call(ctx, x, y, w, h);
    const data = img.data;

    const blockSize = 4 + (prng() * 6 | 0); // 4~10 像素
    const shiftRange = 4; // 随机位移最大像素

    // === block 缓存 ===
    const blocks = [];
    const bw = (w / blockSize) | 0;
    const bh = (h / blockSize) | 0;

    // 拆 Blocks
    for (let by = 0; by < bh; by++) {
        for (let bx = 0; bx < bw; bx++) {
            const block = [];
            for (let yy = 0; yy < blockSize; yy++) {
                for (let xx = 0; xx < blockSize; xx++) {
                    const px = (by * blockSize + yy) * w + (bx * blockSize + xx);
                    const base = px * 4;
                    block.push([
                        data[base], data[base + 1], data[base + 2], data[base + 3]
                    ]);
                }
            }
            blocks.push(block);
        }
    }

    // === 随机重映射 Block ===
    for (let i = blocks.length - 1; i > 0; i--) {
        const j = (prng() * (i + 1)) | 0;
        const tmp = blocks[i];
        blocks[i] = blocks[j];
        blocks[j] = tmp;
    }

    // === 写回新的 blocks ===
    let index = 0;
    for (let by = 0; by < bh; by++) {
        for (let bx = 0; bx < bw; bx++) {
            const block = blocks[index++];

            const offsetX = ((prng() * shiftRange) | 0) - (shiftRange / 2 | 0);
            const offsetY = ((prng() * shiftRange) | 0) - (shiftRange / 2 | 0);

            let i2 = 0;
            for (let yy = 0; yy < blockSize; yy++) {
                for (let xx = 0; xx < blockSize; xx++) {
                    const sx = xx + offsetX;
                    const sy = yy + offsetY;

                    if (sx < 0 || sx >= blockSize || sy < 0 || sy >= blockSize) {
                        i2++;
                        continue;
                    }

                    const px = (by * blockSize + yy) * w + (bx * blockSize + xx);
                    const base = px * 4;

                    let [r, g, b, a] = block[sy * blockSize + sx];

                    // === 非线性颜色变换 ===
                    r = Math.pow(r / 255, 0.8 + prng() * 0.4) * 255;
                    g = Math.pow(g / 255, 0.8 + prng() * 0.4) * 255;
                    b = Math.pow(b / 255, 0.8 + prng() * 0.4) * 255;

                    data[base] = r;
                    data[base + 1] = g;
                    data[base + 2] = b;
                    data[base + 3] = a;

                    i2++;
                }
            }
        }
    }
    return img;
}











 {
    onEnable: ({ win, conf, useSeed, useProxy, useGetterProxy }) => {
    
      const mem = new WeakSet()
      useProxy(win.AudioBuffer.prototype, 'getChannelData', {
        apply: (target, thisArg: AudioBuffer, args: Parameters<typeof AudioBuffer.prototype.getChannelData>) => {
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
      })

      useProxy(win.AudioBuffer.prototype, [
        'copyFromChannel', 'copyToChannel',
      ], {
        apply: (target, thisArg: AudioBuffer, args: any) => {
          const channel = args[1]
          if (channel != null) {
            thisArg.getChannelData(channel)
          }
          return target.apply(thisArg, args)
        }
      })

      const dcNoise = seededRandom(seed) * 1e-7;
      useGetterProxy(win.DynamicsCompressorNode.prototype, 'reduction', (_, getter) => ({
        apply(target, thisArg, args: any) {
          const res = getter.call(thisArg);
          return (typeof res === 'number' && res !== 0) ? res + dcNoise : res;
        }
      }))

    },
  },