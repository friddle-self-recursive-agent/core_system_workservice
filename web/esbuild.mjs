/* esbuild.mjs —— 构建脚本：node esbuild.mjs → dist/console.js + dist/console.css。
 * bundle 形态=IIFE + globalName=CoreConsoleWeb（dsh plugin 打包器消费——零模块加载器假设，
 * 不产 importmap/ESM chunk）。CSS 经 JS 入口 import 三件样式，esbuild 自动抽取为同名
 * .css 产物（与 console.js 同基名）。minify 关闭——首版产物可读、便于外科手术式排查。 */
import { build } from 'esbuild';
import { mkdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.dirname(fileURLToPath(import.meta.url));
mkdirSync(path.join(root, 'dist'), { recursive: true });

const res = await build({
  entryPoints: [path.join(root, 'src/main.js')],
  outfile: path.join(root, 'dist/console.js'),
  bundle: true,
  format: 'iife',
  globalName: 'CoreConsoleWeb',
  platform: 'browser',
  target: ['es2020'],
  sourcemap: false,
  minify: false,
  logLevel: 'info',
  legalComments: 'none',
});

if (res.errors && res.errors.length) {
  console.error('[esbuild] 构建失败');
  process.exit(1);
}
console.log('[esbuild] 产物就绪：dist/console.js + dist/console.css');
