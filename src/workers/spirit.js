// 灵力获取 Worker
// 由于灵力获取已经移到后端处理，这里只需定期同步数据

let intervalId = null;

self.onmessage = function(e) {
  if (e.data.type === 'start') {
    // 启动定时器，定期从后端同步数据
    intervalId = setInterval(() => {
      // 发送消息通知主线程需要同步数据
      self.postMessage({ type: 'sync' });
    }, 2000); // 每10秒同步一次数据，提高同步频率
  } else if (e.data.type === 'stop') {
    // 停止定时器
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
  }
};
