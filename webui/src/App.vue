<script setup>
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { apiLogout } from '@/services/api.js';

// Ottiene l'oggetto 'route' corrente
const route = useRoute();

// Crea una proprietà reattiva che è 'true' se siamo sulla pagina di Login
const isLoginPage = computed(() => route.name === 'Login');

const handleLogout = () => {
  apiLogout();
};
</script>

<template>

  <!-- MOSTRA QUESTO BLOCCO (IL LAYOUT COMPLETO) SOLO SE NON SIAMO SULLA PAGINA DI LOGIN -->
  <template v-if="!isLoginPage">
    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
      
      <!-- Pulsante Logout -->
      <div class="navbar-nav">
        <div class="nav-item text-nowrap">
          <a class="nav-link px-3" href="#" @click.prevent="handleLogout">
            Logout
            <svg class="feather" style="width: 24px; height: 24px; vertical-align: middle; margin-left: 5px;"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
          </a>
        </div>
      </div>
      
      <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
    </header>

    <div class="container-fluid">
      <div class="row">
        <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
          <div class="position-sticky pt-3 sidebar-sticky">
            
            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
              <span>Conversazioni</span>
            </h6>
            <ul class="nav flex-column">
              <li class="nav-item">
                <RouterLink to="/" class="nav-link" active-class="active">
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#home"/></svg>
                  Home
                </RouterLink>
              </li>
              <!-- Aggiungeremo qui 'Cerca Utenti' e 'Settings' -->
            </ul>

          </div>
        </nav>

        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <!-- <RouterView /> carica 'Home.vue' o 'Chat.vue' qui dentro -->
          <RouterView />
        </main>
      </div>
    </div>
  </template>

  <!-- MOSTRA QUESTO BLOCCO SOLO SE SIAMO SULLA PAGINA DI LOGIN -->
  <template v-else>
    <!-- <RouterView /> carica 'Login.vue' qui, a schermo intero senza layout -->
    <RouterView />
  </template>

</template>

<style>
/* Stili globali per far funzionare bene i link attivi e le icone */
.feather {
  width: 16px;
  height: 16px;
  vertical-align: text-bottom;
  margin-right: 8px;
}

.sidebar .nav-link.active {
  color: #2470dc;
  font-weight: 500;
}
</style>