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
    
    // --- CORREZIONE IMPORTANTE ---
    // Il nostro backend (correttamente) restituisce { "users": [...] }
    // Dobbiamo estrarre l'array 'data.users'
    searchResults.value = data.users || []; 
    
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

    // 1. Chiama l'API e ASPETTA la risposta
    const newConversation = await apiStartConversation(userId);
    
    // 2. Successo! Avvisa il genitore (Home.vue)
    //    e passagli l'ID della chat appena creata.
    emit('chat-created', newConversation.id);
    
    
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
        <button type="button" class="btn-close" aria-label="Close" @click="emit('close')" />
      </div>
      
      <div class="card-body">
        <!-- Form di Ricerca -->
        <form @submit.prevent="handleSearch">
          <div class="input-group mb-3">
            <input 
              v-model="searchQuery" 
              type="text" 
              class="form-control" 
              placeholder="Nome utente..."
              required
            >
            <button class="btn btn-outline-primary" type="submit" :disabled="loading">
              <span v-if="loading" class="spinner-border spinner-border-sm" role="status" aria-hidden="true" />
              <svg v-else class="feather"><use href="/feather-sprite-v4.29.0.svg#search" /></svg>
            </button>
          </div>
        </form>

        <!-- Area Risultati -->
        <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

        <div v-if="searchResults.length > 0" class="list-group list-group-flush">
          <a 
            v-for="user in searchResults"
            :key="user.id" 
            href="#"
            class="list-group-item list-group-item-action d-flex gap-3 py-2 align-items-center"
            @click.prevent="handleStartChat(user.id)"
          >
            
            <img
              :src="user.photoUrl || 'https://placehold.co/40x40/e9ecef/000000?text=' + user.username.charAt(0).toUpperCase()" 
              alt="foto" width="40" height="40" class="rounded-circle flex-shrink-0 border"
              style="object-fit: cover;"
            >
            
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
  z-index: 1050;
}

.modal-content {
  width: 100%;
  max-width: 500px;
  margin: 1rem;
  background-color: white;
  border-radius: 0.375rem; /* Aggiunge i bordi arrotondati del 'card' */
}
</style>