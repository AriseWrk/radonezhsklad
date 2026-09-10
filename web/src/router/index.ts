import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
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
      { path: 'products',   name: 'products',   component: () => import('../views/ProductsView.vue') },
      { path: 'categories', name: 'categories', component: () => import('../views/CategoriesView.vue') },
      { path: 'warehouses', name: 'warehouses', component: () => import('../views/WarehousesView.vue') },
      { path: 'stock',      name: 'stock',      component: () => import('../views/StockView.vue') },
      { path: 'documents',  name: 'documents',  component: () => import('../views/DocumentsView.vue') },
      { path: 'customers',  name: 'customers',  component: () => import('../views/CustomersView.vue') },
      { path: 'orders',     name: 'orders',     component: () => import('../views/OrdersView.vue') },
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
})