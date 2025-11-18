<script setup>

import { computed } from 'vue'; // Importa la funzione computed di Vue per le proprietà calcolate
import { RouterLink, RouterView, useRoute } from 'vue-router'; // Importa componenti e funzioni di routing
import { apiLogout } from '@/services/api.js'; // Importa la funzione per il logout dall'API

// Ottiene l'oggetto route corrente
const route = useRoute();

// Determina se la pagina corrente è la pagina di Login
const isLoginPage = computed(() => route.name === 'Login');

// Funzione per gestire il logout
const handleLogout = () => {
  apiLogout();
};

</script>

<template> 

  <!-- Layout principale, visibile solo se non siamo sulla pagina di Login -->

  <template v-if="!isLoginPage"> <!-- Controlla che non siamo sulla pagina di Login -->

    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow"> <!-- Barra di navigazione superiore -->
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a> <!-- Logo e nome dell'app -->
      
      <!-- Pulsante Logout -->
      <div class="navbar-nav">
        <div class="nav-item text-nowrap">
          <a class="nav-link px-3" href="#" @click.prevent="handleLogout">
            Logout
            <svg class="feather" style="width: 24px; height: 24px; vertical-align: middle; margin-left: 5px;"><use href="/feather-sprite-v4.29.0.svg#log-out" /></svg>
          </a>
        </div>
      </div>

      <!-- Pulsante per nascondere la sidebar -->
      <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
    </header>

    <!-- Contenuto principale con sidebar e area di visualizzazione -->
    <div class="container-fluid">
      <div class="row">
        <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
          <div class="position-sticky pt-3 sidebar-sticky">
            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
              <span>Conversazioni</span>
            </h6>
            <ul class="nav flex-column">
              <li class="nav-item">  
                <RouterLink to="/" class="nav-link" active-class="active"> <!-- Link alla Home -->
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#home" /></svg> 
                  Home
                </RouterLink>
              </li>
            </ul>
          </div>
        </nav>

        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <!-- <RouterView /> carica i componenti delle pagine qui -->
          <RouterView />  
        </main>
      </div>
    </div>
  </template>

  <!-- Pagina di Login, senza layout -->
  <template v-else>
    <!-- <RouterView /> carica 'Login.vue' qui -->
    <RouterView />
  </template>
</template>

<style>
/* Stili globali */

/* Stile per le icone SVG */
.feather { 
  width: 16px; /* Larghezza dell'icona */
  height: 16px; /* Altezza dell'icona */
  vertical-align: text-bottom; /* Allineamento verticale */
  margin-right: 8px; /* Spazio a destra dell'icona */
}

/* Stile per il link attivo nella sidebar */
.sidebar .nav-link.active {  
  color: #2470dc; /* Colore blu per il link attivo */
  font-weight: 500; /* Testo in grassetto per il link attivo */
}
</style>