import { createRouter, createWebHistory } from 'vue-router';
// Importa le Viste che abbiamo appena creato
import Home from '@/views/Home.vue';
import Login from '@/views/Login.vue';

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home,
    // (In futuro, potremmo aggiungere 'meta: { requiresAuth: true }' 
    // ma per ora il guard copre tutto)
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  // { path: '/conversations/:id', name: 'Chat', component: () => import('@/views/Chat.vue') },
  // { path: '/settings', name: 'Settings', component: () => import('@/views/Settings.vue') },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

/**
 * Navigation Guard (Il "Controllore")
 * Questo blocco viene eseguito PRIMA di ogni cambio di pagina.
 */
router.beforeEach((to, from, next) => {
  // Controlla se l'utente ha un token nel localStorage
  const isLoggedIn = !!localStorage.getItem('sessionToken');
  
  if (to.name !== 'Login' && !isLoggedIn) {
    // Se l'utente NON è loggato E sta cercando di andare
    // in qualsiasi pagina tranne 'Login',
    // forzalo alla pagina di Login.
    next({ name: 'Login' });
  } else if (to.name === 'Login' && isLoggedIn) {
    // Se l'utente È loggato E sta cercando di andare
    // alla pagina 'Login' (magari scrivendo l'URL a mano),
    // forzalo alla pagina 'Home'.
    next({ name: 'Home' });
  } else {
    // In tutti gli altri casi (sei loggato e vai alla Home, 
    // o non sei loggato e vai al Login), lascia che proceda.
    next();
  }
});

export default router;