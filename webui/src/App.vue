<script setup>
import { ref, computed, watch } from 'vue'; // <--- Aggiungi 'watch' qui
import { RouterLink, RouterView, useRoute } from 'vue-router'; 
import { apiLogout } from '@/services/api.js'; 
import ProfileModal from '@/components/ProfileModal.vue'; 

const route = useRoute();
const isLoginPage = computed(() => route.name === 'Login');

const showProfileModal = ref(false);
const currentUsername = ref(localStorage.getItem('username') || 'Utente');

const handleLogout = () => {
  apiLogout();
};

const onProfileUpdated = (newName) => {
  currentUsername.value = newName;
};

// --- FIX DEL NOME "UTENTE" ---
// Osserviamo la rotta: ogni volta che l'utente cambia pagina (es. da Login a Home),
// rileggiamo il nome dal localStorage per essere sicuri di avere quello aggiornato.
watch(
  () => route.name,
  () => {
    const savedName = localStorage.getItem('username');
    if (savedName) {
      currentUsername.value = savedName;
    }
  }
);
</script>

<template> 
  <template v-if="!isLoginPage"> 

    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
      
      <div class="navbar-nav ms-auto d-flex flex-row align-items-center"> 
        
        <div class="nav-item text-nowrap me-3">
          <a class="nav-link px-3 d-flex align-items-center" href="#" @click.prevent="showProfileModal = true">
            <span class="me-2">{{ currentUsername }}</span>
            <svg class="feather" style="width: 20px; height: 20px; margin: 0;"><use href="/feather-sprite-v4.29.0.svg#user" /></svg>
          </a>
        </div>

        <div class="nav-item text-nowrap me-3">
          <a class="nav-link px-3" href="#" @click.prevent="handleLogout" title="Esci">
            <svg class="feather" style="width: 20px; height: 20px; margin: 0;"><use href="/feather-sprite-v4.29.0.svg#log-out" /></svg>
          </a>
        </div>

      </div>
    </header>

    <div class="container-fluid">
      <div class="row">
        <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
          <div class="position-sticky pt-3 sidebar-sticky">
            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
              <span>Menu</span>
            </h6>
            <ul class="nav flex-column">
              <li class="nav-item">  
                <RouterLink to="/" class="nav-link" active-class="active"> 
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-square" /></svg> 
                  Conversazioni
                </RouterLink>
              </li>
            </ul>
          </div>
        </nav>

        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <RouterView />  
        </main>
      </div>
    </div>

    <ProfileModal 
      v-if="showProfileModal"
      :show="showProfileModal"
      :username="currentUsername"
      @close="showProfileModal = false"
      @update-profile="onProfileUpdated"
    />

  </template>

  <template v-else>
    <RouterView />
  </template>
</template>

<style>
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
/* Fix per navbar mobile */
@media (max-width: 767.98px) {
  #sidebarMenu.collapse.show {
    position: fixed;
    top: 48px; 
    left: 0; right: 0; bottom: 0;
    z-index: 1000; 
    background-color: #f8f9fa; 
    padding-top: 1rem;
    overflow-y: auto; 
    height: calc(100vh - 48px); 
    border-bottom: 1px solid #ddd;
  }
}
</style>