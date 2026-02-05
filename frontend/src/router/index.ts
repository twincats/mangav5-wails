import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'

/**
 * Route meta typing
 */
declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    requiresAuth?: boolean
  }
}

/**
 * Routes
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/HomeView.vue'),
    meta: {
      title: 'Home',
    },
  },
  {
    path: '/chapters/:mangaId',
    name: 'chapters',
    component: () => import('../views/ChapterView.vue'),
    props: route => {
      const raw = Array.isArray(route.params.mangaId)
        ? route.params.mangaId[0]
        : route.params.mangaId

      return { mangaId: Number(raw) }
    },
  },
  {
    path: '/read/:chapterId',
    name: 'read',
    component: () => import('../views/Reader.vue'),
    props: route => {
      const raw = Array.isArray(route.params.chapterId)
        ? route.params.chapterId[0]
        : route.params.chapterId

      return { chapterId: Number(raw) }
    },
  },
  {
    path: '/download',
    name: 'download',
    component: () => import('../views/DownloadView.vue'),
    meta: {
      title: 'Download',
    },
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('../views/SettingsView.vue'),
    meta: {
      title: 'Settings',
      requiresAuth: true,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/NotFoundView.vue'),
    meta: {
      title: 'Not Found',
    },
  },
]

/**
 * Router instance
 *
 * ⚠️ IMPORTANT for Tauri:
 * Use hash history to avoid filesystem routing issues
 */
const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

/**
 * Global navigation guard
 */
router.beforeEach((to, _from, next) => {
  // update window title
  if (to.meta.title) {
    document.title = `My Tauri App – ${to.meta.title}`
  }

  // example auth guard
  if (to.meta.requiresAuth) {
    const isAuthenticated = true // replace with real check
    if (!isAuthenticated) {
      return next({ name: 'home' })
    }
  }

  next()
})

export default router
