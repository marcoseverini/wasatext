<script setup>
import { ref, onMounted } from 'vue'; // Importa onMounted per il ciclo di vita
import { useRouter } from 'vue-router'; // Per il routing
import { apiGetMyConversations } from '@/services/api.js'; // Importa la funzione API per ottenere le conversazioni
import ErrorMsg from '@/components/ErrorMsg.vue'; // Componente per mostrare messaggi di errore
import LoadingSpinner from '@/components/LoadingSpinner.vue'; // Componente per mostrare uno spinner di caricamento
import SearchModal from '@/components/SearchModal.vue'; // Componente per il modale di ricerca utenti
import CreateGroupModal from '@/components/CreateGroupModal.vue'; // Componente per il modale di creazione gruppo

// Definiamo le variabili reattive
const conversations = ref([]); // Lista delle conversazioni
const loading = ref(true); // Stato di caricamento
const errorMsg = ref(''); // Messaggio di errore
const router = useRouter(); // Ottiene l'istanza del router
const isSearchModalVisible = ref(false); // Variabile per il modale di ricerca utenti

// Aggiunge la variabile per il nuovo modale
const isCreateGroupModalVisible = ref(false);

// Funzione per caricare le conversazioni
const loadConversations = async () => {
  try {
    loading.value = true;
    errorMsg.value = '';
    const data = await apiGetMyConversations();
    conversations.value = data.conversations || [];
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};

// Funzione per andare alla chat
const goToConversation = (convId) => {
  router.push(`/conversations/${convId}`);
};

// Funzione da chiamare quando il modale di ricerca ha creato una chat
const onChatCreated = (newConvId) => {
  isSearchModalVisible.value = false; // Chiude il modale
  router.push(`/conversations/${newConvId}`); // Naviga alla nuova chat
};

// Aggiunge la funzione per il modale del gruppo
const onGroupCreated = (newGroupId) => {
  isCreateGroupModalVisible.value = false; // Chiudi il modale
  router.push(`/conversations/${newGroupId}`); // Naviga alla nuova chat
};

// Carica le conversazioni al montaggio della pagina
onMounted(() => {
  loadConversations();
});
</script>

<template>
  <div>
    <!-- Intestazione della Pagina -->
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Le mie Conversazioni</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <button type="button" class="btn btn-sm btn-outline-secondary me-2" @click="isSearchModalVisible = true">
          <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search" /></svg>
          Cerca Utenti
        </button>
        <!-- Pulsante per aprire il modale di creazione gruppo -->
        <button type="button" class="btn btn-sm btn-outline-primary" @click="isCreateGroupModalVisible = true">
          <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#users" /></svg>
          Crea Gruppo
        </button>
      </div>
    </div>

    <!-- Messaggio di Errore -->
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <!-- Spinner di Caricamento -->
    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
      <p>Caricamento conversazioni...</p>
    </div>

    <!-- Lista Conversazioni -->
    <div v-if="!loading && !errorMsg">
      <!-- Stato Vuoto -->
      <div v-if="conversations.length === 0" class="text-center text-muted mt-5">
        <p>Non hai ancora nessuna conversazione.</p>
        <p>Usa "Cerca Utenti" o "Crea Gruppo" per iniziarne una!</p>
      </div>

      <!-- La Lista (questa parte è invariata) -->
      <div v-else class="list-group">
        <a 
          v-for="convo in conversations"
          :key="convo.id" 
          href="#"
          class="list-group-item list-group-item-action d-flex gap-3 py-3"
          @click.prevent="goToConversation(convo.id)"
        >
          
          <img
            :src="convo.photoUrl || 'https://placehold.co/64x64/25d366/FFF?text=' + convo.name.charAt(0)" 
            alt="foto" width="64" height="64" class="rounded-circle flex-shrink-0"
          >
          
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

    <!-- Modale Ricerca Utenti -->
    <SearchModal 
      :show="isSearchModalVisible" 
      @close="isSearchModalVisible = false"
      @chat-created="onChatCreated"
    />

    <!-- Aggiunge il nuovo modale per creare i gruppi -->
    <CreateGroupModal
      :show="isCreateGroupModalVisible"
      @close="isCreateGroupModalVisible = false"
      @group-created="onGroupCreated"
    />
  </div>
</template>

<style> /* Stili specifici per la Home.vue */

/* Stile per gli elementi della lista delle conversazioni */
.list-group-item-action { 
  align-items: center;
}

/* Stile per il testo che potrebbe essere troppo lungo */
.list-group-item-action .mb-0 {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 40vw; 
}

</style>