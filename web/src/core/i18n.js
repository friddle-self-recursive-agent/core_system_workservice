/* i18n.js —— 控制台双语层（默认中文源词形，可切英文；字符串键=中文源词形，缺译回退键本身——
 * 诚实降级不冒充翻译）。locale 持久化 localStorage('cc.locale')；切换经 subscribers 通知
 * （main.js 重绘侧栏壳+当前屏；屏内串经 t() 在渲染期解析——重绘即换语言，零热更新 hack）。
 * EN 词典本体在 ./i18n-en.js（R-ENG-7 拆分——词典是数据不是逻辑）。 */

import { enEntries } from './i18n-en.js';

const KEY = 'cc.locale';
let locale = 'zh';
try {
  locale = (typeof localStorage !== 'undefined' && localStorage.getItem(KEY)) || 'zh';
} catch { /* 隐私模式/无 localStorage=内存态 */ }

/* EN 词典：键=中文源词形，值=英文词形（本体在 i18n-en.js）。 */
const en = new Map(enEntries);

const DICTS = { en };
const subs = new Set();

/** 当前词形解析：locale=zh（源语言）恒返回原文；en 查词典，缺译回退键本身。 */
export function t(s) {
  if (locale === 'zh') return String(s);
  const d = DICTS[locale];
  return (d && d.get(String(s))) || String(s);
}

export function getLocale() { return locale; }

export function setLocale(l) {
  if (l !== 'zh' && !DICTS[l]) return; // 未注册词形=忽略（不炸运行态）
  locale = l;
  try { if (typeof localStorage !== 'undefined') localStorage.setItem(KEY, l); } catch { /* ignore */ }
  subs.forEach((f) => { try { f(l); } catch { /* 订阅者自兜底 */ } });
}

export function onLocaleChange(f) {
  subs.add(f);
  return () => subs.delete(f);
}

/** 屏内 t() 键清单（测试对照 EN 词典覆盖度用——缺译计数=诚实账）。 */
export function dictOf(l) { return DICTS[l] || new Map(); }
