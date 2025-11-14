<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { apiGetMyConversations } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Variabili reattive
const conversations = ref([]); // La lista delle chat
const loading = ref(true);
const errorMsg = ref('');
const router = useRouter(); // Per navigare

// Funzione per caricare le conversazioni
const loadConversations = async () => {
  try {
    loading.value = true;
    errorMsg.value = '';
    
    // Chiama l'API che abbiamo definito in api.js
    const data = await apiGetMyConversations();
    
    // Salva solo l'array
    conversations.value = data.conversations || [];
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};

// Funzione (per ora vuota) per andare alla chat
const goToConversation = (convId) => {
  // In futuro: router.push(`/conversations/${convId}`);
  console.log("Andiamo alla chat:", convId);
};

// Funzione (per ora vuota) per mostrare il modale di ricerca
const showSearchModal = () => {
  console.log("Apri modale ricerca utenti");
  // Qui implementeremo il modale come Lachi
};

// Funzione (per ora vuota) per mostrare il modale di creazione gruppo
const showCreateGroupModal = () => {
  console.log("Apri modale crea gruppo");
  // Qui implementeremo il modale
};

// onMounted() viene eseguito quando il componente viene caricato
onMounted(() => {
  loadConversations();
});
</script>

<template>
  <div>
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Le mie Conversazioni</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <button type="button" class="btn btn-sm btn-outline-secondary me-2" @click="showSearchModal">
          <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
          Cerca Utenti
        </button>
        <button type="button" class="btn btn-sm btn-outline-primary" @click="showCreateGroupModal">
          <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
          Crea Gruppo
        </button>
      </div>
    </div>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
      <p>Caricamento conversazioni...</p>
    </div>

    <div v-if="!loading && !errorMsg">
      
      <div v-if="conversations.length === 0" class="text-center text-muted mt-5">
        <p>Non hai ancora nessuna conversazione.</p>
        <p>Usa "Cerca Utenti" per iniziarne una!</p>
      </div>

      <div v-else class="list-group">
        <a 
          href="#"
          v-for="convo in conversations" 
          :key="convo.id"
          @click.prevent="goToConversation(convo.id)"
          class="list-group-item list-group-item-action d-flex gap-3 py-3">
          
          <img :src="convo.photoUrl || 'https://placehold.co/64x64/25d366/FFF?text=' + convo.name.charAt(0)" 
               alt="foto" width="64" height="64" class="rounded-circle flex-shrink-0">
          
          <div class="d-flex gap-2 w-100 justify-content-between">
            <div>
              <h6 class="mb-0">{{ convo.name }}</h6>
              <p class="mb-0 opacity-75">{{ convo.latestMessageSnippet || 'Nessun messaggio' }}</p>
            </div>
            <small class="opacity-50 text-nowrap">{{ convo.latestMessageTimestamp }}</small>
          </div>
        </a>
      </div>

    </div>

  </div>
</template>

<style>
/* Stili per rendere la lista più simile a un'app di chat */
.list-group-item-action {
  align-items: center;
}
.list-group-item-action .mb-0 {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 40vw; /* Impedisce al testo di andare a capo */
}
</style>