// 错误提示去重：并发请求同时失败时，避免弹出多个相同提示。
const WINDOW = 3000; // 3 秒窗口

export function createErrorDedup(windowMs = WINDOW) {
  let lastNetworkAt = 0;
  let lastBizKey = '';
  let lastBizAt = 0;
  return {
    /** 网络异常：窗口期内只提示一次 */
    shouldShowNetwork(): boolean {
      const now = Date.now();
      if (now - lastNetworkAt < windowMs) return false;
      lastNetworkAt = now;
      return true;
    },
    /** 业务错误：相同文案窗口期内只提示一次 */
    shouldShowBiz(msg: string): boolean {
      const now = Date.now();
      if (lastBizKey === msg && now - lastBizAt < windowMs) return false;
      lastBizKey = msg;
      lastBizAt = now;
      return true;
    },
  };
}
