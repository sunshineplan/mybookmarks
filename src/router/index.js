import { createRouter, createWebHistory } from 'vue-router'


const routes = [
  {
    path: '/',
    component: () => import(/* webpackChunkName: 'show' */ '../views/ShowBookmarks.vue')
  },
  {
    path: '/setting',
    component: () => import(/* webpackChunkName: 'setting' */ '../views/Setting.vue')
  },
  {
    path: '/category/:mode',
    component: () => import(/* webpackChunkName: 'category' */ '../views/Category.vue')
  },
  {
    path: '/bookmark/:mode',
    component: () => import(/* webpackChunkName: 'bookmark' */ '../views/Bookmark.vue')
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
