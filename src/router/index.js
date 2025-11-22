import { createRouter, createWebHashHistory } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import Home from '../views/Home.vue'
import Cultivation from '../views/Cultivation.vue'
import Inventory from '../views/Inventory.vue'
import Exploration from '../views/Exploration.vue'
import Settings from '../views/Settings.vue'
import Alchemy from '../views/Alchemy.vue'
import Dungeon from '../views/Dungeon.vue'
import Gacha from '../views/Gacha.vue'
import Leaderboard from '../views/Leaderboard.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/cultivation',
    name: 'Cultivation',
    component: Cultivation
  },
  {
    path: '/inventory',
    name: 'Inventory',
    component: Inventory
  },
  {
    path: '/exploration',
    name: 'Exploration',
    component: Exploration
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings
  },
  {
    path: '/alchemy',
    name: 'alchemy',
    component: Alchemy
  },
  {
    path: '/dungeon',
    name: 'Dungeon',
    component: Dungeon
  },
  {
    path: '/gacha',
    name: 'Gacha',
    component: Gacha
  },
  {
    path: '/leaderboard',
    name: 'Leaderboard',
    component: Leaderboard
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router