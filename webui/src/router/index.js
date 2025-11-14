import { createRouter, createWebHistory } from 'vue-router';

import Home from '@/views/Home.vue';
import Login from '@/views/Login.vue';
import Chat from '@/views/Chat.vue';

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home,
    // In futuro, potremmo aggiungere 'meta: { requiresAuth: true }' ma per ora il guard copre tutto
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    // :id è un parametro dinamico (es. /conversations/conv-123)
    path: '/conversations/:id', 
    name: 'Chat',
    component: Chat,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Questo blocco viene eseguito prima di ogni cambio di pagina
router.beforeEach((to, from, next) => {

  // Controlla se l'utente ha un token nel localStorage
  const isLoggedIn = !!localStorage.getItem('sessionToken');
  
  if (to.name !== 'Login' && !isLoggedIn) {
    // Se l'utente non è loggato e 
    // sta cercando di andare in qualsiasi pagina tranne 'Login',
    // viene rimandato alla pagina di Login.
    next({ name: 'Login' });
  } else if (to.name === 'Login' && isLoggedIn) {
    // Se l'utente è loggato e sta cercando di andare alla pagina 'Login',
    // viene rimandato alla pagina 'Home'.
    next({ name: 'Home' });
  } else {
    // In tutti gli altri casi, lascia che proceda.
    next();
  }
});

export default router;