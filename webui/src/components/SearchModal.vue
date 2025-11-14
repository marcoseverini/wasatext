<script setup>
import { ref } from 'vue';
import { apiSearchUsers, apiStartConversation } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Definiamo i 'props' (dati in ingresso) e 'emits' (eventi in uscita)
// Il genitore (Home.vue) usa 'show' per aprirlo
// e ascolta '@close' e '@chat-created'
defineProps({
  show: Boolean
});
const emit = defineEmits(['close', 'chat-created']);

// Variabili reattive per il modale
const searchQuery = ref('');
const loading = ref(false);
const errorMsg = ref('');
const searchResults = ref([]); // Lista dei risultati

// Funzione per eseguire la ricerca
const handleSearch = async () => {
  if (searchQuery.value.length < 1) {
    errorMsg.value = "Inserisci almeno 1 carattere per cercare.";
    return;
  }
  
  loading.value = true;
  errorMsg.value = '';
  searchResults.value = [];
  
  try {
    // Chiama l'API
    const data = await apiSearchUsers(searchQuery.value);
    // Ricorda che /users restituisce un array nudo, non un oggetto
    searchResults.value = data || []; 
    
    if (searchResults.value.length === 0) {
      errorMsg.value = "Nessun utente trovato.";
    }
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};

// Funzione per avviare la chat quando si clicca un utente
const handleStartChat = async (userId) => {
  loading.value = true;
  errorMsg.value = '';
  
  try {
    // Chiama l'API per creare la chat 1-a-1
    await apiStartConversation(userId);
    
    // Successo! Avvisa il genitore (Home.vue)
    emit('chat-created');
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <!-- 
    'v-if' è cruciale. Il componente non esiste nel DOM
    finché 'show' non è 'true'.
  -->
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="modal-title mb-0">Cerca un Utente</h5>
        <button type="button" class="btn-close" @click="emit('close')" aria-label="Close"></button>
      </div>
      
      <div class="card-body">
        <!-- Form di Ricerca -->
        <form @submit.prevent="handleSearch">
          <div class="input-group mb-3">
            <input 
              type="text" 
              class="form-control" 
              placeholder="Nome utente..." 
              v-model="searchQuery"
              required>
            <button class="btn btn-outline-primary" type="submit" :disabled="loading">
              <span v-if="loading" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
              <svg v-else class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
            </button>
          </div>
        </form>

        <!-- Area Risultati -->
        <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

        <div v-if="searchResults.length > 0" class="list-group list-group-flush">
          <a 
            href="#"
            v-for="user in searchResults" 
            :key="user.id"
            @click.prevent="handleStartChat(user.id)"
            class="list-group-item list-group-item-action d-flex gap-3 py-2 align-items-center">
            
            <img :src="user.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + user.username.charAt(0)" 
                 alt="foto" width="40" height="40" class="rounded-circle flex-shrink-0">
            
            <h6 class="mb-0">{{ user.username }}</h6>
          </a>
        </div>
        
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Stili per il modale (pop-up) */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1050; /* Sopra la navbar di bootstrap */
}

.modal-content {
  width: 100%;
  max-width: 500px;
  margin: 1rem;
}
</style>