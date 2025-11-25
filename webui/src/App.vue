<script setup>
import { ref, computed, watch, onMounted } from 'vue'; 
import { RouterLink, RouterView, useRoute } from 'vue-router'; 
import { apiLogout, apiSearchUsers } from '@/services/api.js'; // Importa apiSearchUsers
import ProfileModal from '@/components/ProfileModal.vue'; 

const route = useRoute();
const isLoginPage = computed(() => route.name === 'Login');

const showProfileModal = ref(false);

// Stato Utente
const currentUsername = ref(localStorage.getItem('username') || 'Utente');
const currentPhotoUrl = ref(localStorage.getItem('photoUrl') || ''); // Carica foto da localStorage

const handleLogout = () => {
  apiLogout();
};

// Aggiorna stato e localStorage quando il profilo viene modificato dal modale
const onProfileUpdated = ({ username, photoUrl }) => {
  currentUsername.value = username;
  currentPhotoUrl.value = photoUrl;
  
  localStorage.setItem('username', username);
  if (photoUrl) {
    localStorage.setItem('photoUrl', photoUrl);
  } else {
    localStorage.removeItem('photoUrl');
  }
};

// Funzione per recuperare i propri dati freschi dal server (utile se cambio PC o cancello cache)
const fetchMyProfile = async () => {
  const myId = localStorage.getItem('sessionToken');
  const myName = localStorage.getItem('username');
  
  if (!myId || !myName) return;

  try {
    // Cerco me stesso
    const data = await apiSearchUsers(myName);
    console.log("Cercato utente:", myName, "Risultati:", data); // <--- DEBUG 1

    // Cerco l'utente con il mio ID
    const me = data.users.find(u => u.id === myId);
    console.log("Trovato me stesso?", me); // <--- DEBUG 2
    
    if (me) {
      // Se mi trovo, aggiorno lo stato locale
      currentPhotoUrl.value = me.photoUrl || '';
      localStorage.setItem('photoUrl', currentPhotoUrl.value);
      console.log("Foto aggiornata:", currentPhotoUrl.value); // <--- DEBUG 3
    }
  } catch (e) {
    console.error("Impossibile recuperare profilo utente", e);
  }
};

// Al caricamento dell'App, proviamo a recuperare la foto fresca
onMounted(() => {
  if (!isLoginPage.value) {
    fetchMyProfile();
  }
});

watch(
  () => route.name,
  () => {
    const savedName = localStorage.getItem('username');
    if (savedName) {
      currentUsername.value = savedName;
    } else { // vedi mpo
      // Se non c'è username (es. logout), resetta tutto
      currentUsername.value = 'Utente';
      currentPhotoUrl.value = ''; 
    }
    
    // Aggiorna anche la foto al cambio rotta (es. dopo login)
    const savedPhoto = localStorage.getItem('photoUrl');
    if (savedPhoto) currentPhotoUrl.value = savedPhoto;
    
    // Se siamo loggati, rinfresca i dati dal server
    if (route.name !== 'Login') fetchMyProfile();
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
            
            <img 
              v-if="currentPhotoUrl" 
              :src="currentPhotoUrl" 
              class="rounded-circle border border-secondary"
              width="24" height="24" 
              style="object-fit: cover;"
            >
            <svg v-else class="feather" style="width: 20px; height: 20px; margin: 0;"><use href="/feather-sprite-v4.29.0.svg#user" /></svg>
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
      :photoUrl="currentPhotoUrl"  
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