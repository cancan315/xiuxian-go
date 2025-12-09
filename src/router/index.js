import { createRouter, createWebHashHistory } from 'vue-router'
import Login from '../views/Login.vue'
import { getAuthToken } from '../stores/db'

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/home',
    name: 'App',
    component: () => import('../App.vue'),
    beforeEnter: (to, from, next) => {
      // 检查是否有认证令牌
      if (!getAuthToken()) {
        // 如果没有认证令牌，重定向到登录页
        next('/login')
      } else {
        // 如果有认证令牌，允许访问
        next()
      }
    }
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router