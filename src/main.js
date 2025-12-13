import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import APIService from './services/api'
import { getAuthToken } from './stores/db'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 监听页面卸载事件，主动标记玩家离线
window.addEventListener('beforeunload', async (event) => {
  try {
    const token = getAuthToken();
    if (token) {
      const userData = await APIService.getUser(token);
      if (userData && userData.id) {
        // 直接调用离线接口，避免调用未定义的函数
        await APIService.playerOffline(userData.id);
      }
    }
  } catch (error) {
    console.error('页面卸载时设置离线状态失败:', error);
  }
});

app.mount('#app')
