import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore, type Role } from '../stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    public?: boolean
    roles?: Role[]
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      { path: '',           name: 'dashboard',  component: () => import('../views/DashboardView.vue') },
      { path: 'products',   name: 'products',   component: () => import('../views/ProductsView.vue'),   meta: { roles: ['admin', 'manager', 'warehouse', 'user'] } },
      { path: 'categories', name: 'categories', component: () => import('../views/CategoriesView.vue'), meta: { roles: ['admin', 'manager', 'warehouse', 'user'] } },
      { path: 'warehouses', name: 'warehouses', component: () => import('../views/WarehousesView.vue'), meta: { roles: ['admin', 'warehouse'] } },
      { path: 'stock',      name: 'stock',      component: () => import('../views/StockView.vue'),      meta: { roles: ['admin', 'manager', 'warehouse'] } },
      { path: 'documents',  name: 'documents',  component: () => import('../views/DocumentsView.vue'),  meta: { roles: ['admin', 'manager', 'warehouse'] } },
      { path: 'inventory',  name: 'inventory',  component: () => import('../views/InventoryView.vue'),  meta: { roles: ['admin', 'manager', 'warehouse'] } },
      { path: 'customers',  name: 'customers',  component: () => import('../views/CustomersView.vue'),  meta: { roles: ['admin', 'manager'] } },
      { path: 'orders',     name: 'orders',     component: () => import('../views/OrdersView.vue'),     meta: { roles: ['admin', 'manager'] } },
      { path: 'users',      name: 'users',      component: () => import('../views/UsersView.vue'),      meta: { roles: ['admin'] } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
  if (to.meta.roles && auth.role && !to.meta.roles.includes(auth.role)) {
    return { name: 'dashboard' }
  }
})