<script setup>
import { ref, onMounted } from 'vue';
import { 
  apiGetMyConversations, 
  apiSearchUsers, 
  apiSendMessage, 
  apiStartConversation 
} from '@/services/api.js';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const props = defineProps({
  show: Boolean,
  text: String,
  photo: String
});

const emit = defineEmits(['close', 'forward-success']);

const recentConversations = ref([]);
const searchResults = ref([]);
const searchQuery = ref('');
const loading = ref(false);
const sendingToId = ref(null); 
const hasSearched = ref(false);

onMounted(async () => {
  try {
    loading.value = true;
    const data = await apiGetMyConversations();
    recentConversations.value = data.conversations || [];
  } catch (e) {
    console.error("Errore caricamento chat:", e);
  } finally {
    loading.value = false;
  }
});

const handleSearch = async () => {
  
  if (searchQuery.value.trim().length < 1) return;
  
  loading.value = true;
  hasSearched.value = true; 
  searchResults.value = [];
  
  try {
    const data = await apiSearchUsers(searchQuery.value);
    searchResults.value = data.users || [];
  } catch (e) {
    console.error("Errore ricerca:", e);
  } finally {
    loading.value = false;
  }
};

const handleForward = async (item, isSearchResult) => {
  if (!props.text && !props.photo) return;
  
  sendingToId.value = item.id;
  
  try {
    let targetConvId = item.id;

    if (isSearchResult) {
      const convData = await apiStartConversation(item.id);
      targetConvId = convData.id;
    }

    await apiSendMessage(targetConvId, props.text, props.photo);
    
    alert("Messaggio inoltrato!");
    emit('forward-success');
  } catch (e) {
    alert("Errore nell'inoltro: " + e.message);
  } finally {
    sendingToId.value = null;
  }
};
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content card shadow">
      
      <div class="card-header d-flex justify-content-between align-items-center bg-white border-bottom">
        <h5 class="modal-title mb-0 h6">Inoltra messaggio a...</h5>
        <button type="button" class="btn-close" @click="$emit('close')" />
      </div>
      
      <div class="card-body d-flex flex-column p-3" style="height: 400px;">
        
        <div class="input-group mb-3">
          <input 
            v-model="searchQuery" 
            type="text" 
            class="form-control" 
            placeholder="Cerca utente..." 
            @keyup.enter="handleSearch"
          >
          <button class="btn btn-primary" @click="handleSearch" :disabled="loading">
            Cerca
          </button>
        </div>

        <div class="list-container flex-grow-1 overflow-auto">
          
          <div v-if="loading" class="text-center p-3">
            <LoadingSpinner />
          </div>

          <div v-else>
            
            <div v-if="searchResults.length > 0">
              <h6 class="text-primary small fw-bold text-uppercase mt-2 mb-2 px-1">Risultati Ricerca</h6>
              <div 
                v-for="user in searchResults" 
                :key="user.id" 
                class="d-flex align-items-center justify-content-between p-2 border-bottom user-row"
              >
                <div class="d-flex align-items-center text-truncate">
                  <div class="rounded-circle bg-secondary text-white d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 32px; height: 32px;">
                    {{ user.username.charAt(0).toUpperCase() }}
                  </div>
                  <span class="text-truncate">{{ user.username }}</span>
                </div>
                
                <button class="btn btn-sm btn-outline-primary ms-2" @click="handleForward(user, true)" :disabled="sendingToId !== null">
                  <span v-if="sendingToId === user.id" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                  <span v-else>Invia</span>
                </button>
              </div>
            </div>
            
            <div v-else-if="hasSearched && searchQuery.length >= 1 && searchResults.length === 0" class="text-center text-muted mt-3">
              Nessun utente trovato.
            </div>

            <div v-if="searchResults.length === 0 && recentConversations.length > 0">
              <h6 class="text-muted small fw-bold text-uppercase mt-3 mb-2 px-1">Recenti</h6>
              <div 
                v-for="convo in recentConversations" 
                :key="convo.id"
                class="d-flex align-items-center justify-content-between p-2 border-bottom user-row"
              >
                <div class="d-flex align-items-center text-truncate">
                   <img 
                    :src="convo.photoUrl || 'https://placehold.co/32x32/e9ecef/000000?text=' + convo.name.charAt(0).toUpperCase()" 
                    class="rounded-circle me-2 border flex-shrink-0" width="32" height="32"
                    style="object-fit: cover;"
                   >
                   <span class="text-truncate">{{ convo.name }}</span>
                </div>
                
                <button class="btn btn-sm btn-outline-secondary ms-2" @click="handleForward(convo, false)" :disabled="sendingToId !== null">
                  <span v-if="sendingToId === convo.id" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                  <span v-else>Invia</span>
                </button>
              </div>
            </div>

            <div v-if="recentConversations.length === 0 && searchResults.length === 0 && !hasSearched" class="text-center p-4 text-muted">
              Cerca un utente per iniziare.
            </div>

          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; top: 0; left: 0;
  width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex; justify-content: center; align-items: center;
  z-index: 2000;
}
.modal-content {
  width: 95%; max-width: 450px;
  background-color: white; border-radius: 8px;
}
.user-row:hover {
  background-color: #f8f9fa;
}
</style>