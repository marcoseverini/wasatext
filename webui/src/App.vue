<script setup>
import { ref, computed, watch, onMounted } from 'vue'; 
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'; 
import { apiLogout, apiSearchUsers } from '@/services/api.js'; 

// Importiamo TUTTI i modali qui perché i bottoni sono nella sidebar
import ProfileModal from '@/components/ProfileModal.vue';
import SearchModal from '@/components/SearchModal.vue';
import CreateGroupModal from '@/components/CreateGroupModal.vue';

const route = useRoute();
const router = useRouter();
const isLoginPage = computed(() => route.name === 'Login');

// Stato Utente
const currentUsername = ref(localStorage.getItem('username') || 'Utente');
const currentPhotoUrl = ref(localStorage.getItem('photoUrl') || '');

// Stato Modali
const showProfileModal = ref(false);
const showSearchModal = ref(false);
const showCreateGroupModal = ref(false);

const handleLogout = () => {
  apiLogout();
};

// Aggiornamento Profilo
const onProfileUpdated = ({ username, photoUrl }) => {
  currentUsername.value = username;
  currentPhotoUrl.value = photoUrl;
  localStorage.setItem('username', username);
  if (photoUrl) localStorage.setItem('photoUrl', photoUrl);
  else localStorage.removeItem('photoUrl');
};

// Recupero dati freschi
const fetchMyProfile = async () => {
  const myId = localStorage.getItem('sessionToken');
  const myName = localStorage.getItem('username');
  if (!myId || !myName) return;
  try {
    const data = await apiSearchUsers(myName);
    const me = data.users.find(u => u.id === myId);
    if (me) {
      currentPhotoUrl.value = me.photoUrl || '';
      localStorage.setItem('photoUrl', currentPhotoUrl.value);
    }
  } catch (e) { console.error(e); }
};

// Gestione eventi dai modali globali
const onChatCreated = (newConvId) => {
  showSearchModal.value = false;
  router.push(`/conversations/${newConvId}`);
};

const onGroupCreated = (newGroupId) => {
  showCreateGroupModal.value = false;
  router.push(`/conversations/${newGroupId}`);
};

onMounted(() => {
  if (!isLoginPage.value) fetchMyProfile();
});

watch(() => route.name, () => {
  const savedName = localStorage.getItem('username');
  if (savedName) currentUsername.value = savedName;
  const savedPhoto = localStorage.getItem('photoUrl');
  if (savedPhoto) currentPhotoUrl.value = savedPhoto;
  if (route.name !== 'Login') fetchMyProfile();
});
</script>

<template> 
  <template v-if="!isLoginPage"> 

    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
      <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu">
        <span class="navbar-toggler-icon"></span>
      </button>
    </header>

    <div class="container-fluid">
      <div class="row">
        
        <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
          <div class="position-sticky pt-3 sidebar-sticky d-flex flex-column h-100">
            
            <div class="text-center p-3 border-bottom mb-3">
              <div class="position-relative d-inline-block mb-2">
                <img 
                  :src="currentPhotoUrl || 'https://placehold.co/80x80/e9ecef/6c757d?text=User'" 
                  class="rounded-circle border border-2 border-white shadow-sm"
                  width="80" height="80" 
                  style="object-fit: cover;"
                >
                <button class="btn btn-sm btn-primary position-absolute bottom-0 end-0 rounded-circle p-1 edit-btn" @click="showProfileModal = true" title="Modifica Profilo">
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#edit-2" /></svg>
                </button>
              </div>
              <h6 class="mb-0 fw-bold text-truncate">{{ currentUsername }}</h6>
              <small class="text-muted">Online</small>
            </div>

            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-2 mb-1 text-muted text-uppercase">
              <span>Menu</span>
            </h6>
            <ul class="nav flex-column mb-auto">
              <li class="nav-item">  
                <RouterLink to="/" class="nav-link" active-class="active"> 
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-square" /></svg> 
                  Le mie Chat
                </RouterLink>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#" @click.prevent="showSearchModal = true">
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search" /></svg>
                  Cerca Utenti
                </a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#" @click.prevent="showCreateGroupModal = true">
                  <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#users" /></svg>
                  Nuovo Gruppo
                </a>
              </li>
            </ul>

            <div class="border-top p-3 mt-3">
              <button class="btn btn-outline-danger w-100 d-flex align-items-center justify-content-center gap-2" @click="handleLogout">
                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#log-out" /></svg>
                Esci
              </button>
            </div>

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

    <SearchModal 
      :show="showSearchModal" 
      @close="showSearchModal = false"
      @chat-created="onChatCreated"
    />

    <CreateGroupModal
      :show="showCreateGroupModal"
      @close="showCreateGroupModal = false"
      @group-created="onGroupCreated"
    />

  </template>

  <template v-else>
    <RouterView />
  </template>
</template>

<style>
/* Stili Globali */
.feather { 
  width: 16px; height: 16px; vertical-align: text-bottom; margin-right: 8px; 
}
.sidebar .nav-link {
  color: #333;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  margin: 0 10px;
}
.sidebar .nav-link:hover {
  background-color: #e9ecef;
}
.sidebar .nav-link.active {  
  color: #0d6efd; 
  background-color: #e7f1ff;
  font-weight: 500; 
}
.edit-btn {
  width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;
}
.edit-btn svg { margin: 0; width: 12px; height: 12px; }

/* Sidebar altezza full per posizionare il footer in basso */
.sidebar-sticky {
  height: calc(100vh - 48px);
  overflow-y: auto;
}

@media (max-width: 767.98px) {
  #sidebarMenu.collapse.show {
    position: fixed; top: 48px; left: 0; right: 0; bottom: 0;
    z-index: 1000; background-color: #fff;
  }
}
</style>