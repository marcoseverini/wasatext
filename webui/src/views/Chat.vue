<script setup>
import { ref, onMounted } from 'vue'; // Importa onMounted per il ciclo di vita
import { useRoute } from 'vue-router'; // Per il routing
import { 
  apiGetConversation, 
  apiSendMessage,
  apiDeleteMessage
} from '@/services/api.js'; // Importa le funzioni API necessarie
import ErrorMsg from '@/components/ErrorMsg.vue'; // Componente per mostrare messaggi di errore
import LoadingSpinner from '@/components/LoadingSpinner.vue'; // Componente per mostrare uno spinner di caricamento

// Variabili reattive
const conversation = ref(null); // La conversazione corrente
const loading = ref(true); // Stato di caricamento
const errorMsg = ref(''); // Messaggio di errore
const newMessageText = ref(''); // Testo del nuovo messaggio
const isSending = ref(false); // Stato di invio messaggio

const route = useRoute(); // Ottiene l'istanza della route
const convId = route.params.id; // Ottiene l'ID della conversazione dai parametri della route
const loggedInUserId = localStorage.getItem('sessionToken'); // Ottiene l'ID dell'utente loggato

// Carica la conversazione al montaggio della pagina
onMounted(async () => { 
  try {
    loading.value = true;
    errorMsg.value = '';
    const data = await apiGetConversation(convId);
    if (data.messages) {
      data.messages.reverse(); // Ordina dal più vecchio al più recente
    }
    conversation.value = data;
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
});

// Funzione per inviare un messaggio
const handleSendMessage = async () => {
  if (newMessageText.value.trim() === '') return;
  isSending.value = true;
  errorMsg.value = '';
  try {
    const newMsg = await apiSendMessage(convId, newMessageText.value);
    conversation.value.messages.push(newMsg);
    newMessageText.value = '';
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    isSending.value = false;
  }
};

// Funzione per cancellare il messaggio
const handleDeleteMessage = async (messageId) => {

  // Chiedi conferma
  if (!window.confirm("Sei sicuro di voler cancellare questo messaggio?")) {
    return;
  }

  errorMsg.value = '';
  try {
    // Chiama l'API
    await apiDeleteMessage(messageId);
    
    // Aggiorna la UI: rimuovi il messaggio dalla lista
    conversation.value.messages = conversation.value.messages.filter(
      (msg) => msg.id !== messageId
    );
  } catch (err) {
    errorMsg.value = err.message;
  }
};
</script>

<template>
  <div class="chat-view">
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
    </div>

    <div v-if="!loading && !errorMsg && conversation" class="d-flex flex-column h-100">
      <!-- Intestazione Chat -->
      <div class="d-flex align-items-center pt-3 pb-2 mb-3 border-bottom chat-header">
        <img
          :src="conversation.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + conversation.name.charAt(0)"
          alt="foto" width="40" height="40" class="rounded-circle me-3"
        >
        <h1 class="h4 mb-0">{{ conversation.name }}</h1>
      </div>

      <!-- Area Messaggi -->
      <div class="message-list">
        <div v-if="conversation.messages.length === 0" class="text-center text-muted">
          Questo è l'inizio della tua conversazione.
        </div>
        
        <!-- Itera sui messaggi -->
        <div 
          v-for="msg in conversation.messages" 
          :key="msg.id"
          class="message-wrapper d-flex align-items-center"
          :class="{ 'sent-wrapper': msg.sender.id === loggedInUserId }"
        >
          <!-- Pulsante Cestino (mostrato solo se 'sent') -->
          <button 
            v-if="msg.sender.id === loggedInUserId"
            class="btn btn-sm btn-outline-danger delete-btn"
            @click="handleDeleteMessage(msg.id)"
          >
            <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2" /></svg>
          </button>
          
          <!-- Bolla del Messaggio -->
          <div class="message-bubble" :class="{ 'sent': msg.sender.id === loggedInUserId }"> 
            <div v-if="conversation.isGroup && msg.sender.id !== loggedInUserId" class="message-sender">
              {{ msg.sender.username }}
            </div>
            <div class="message-content">
              {{ msg.content }}
            </div>
            <div class="message-timestamp">
              {{ new Date(msg.timestamp).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' }) }}
            </div>
          </div>
        </div>
      </div>
      
      <!-- Area Scrittura Messaggio -->
      <div class="message-input-area mt-auto">
        <form class="d-flex gap-2" @submit.prevent="handleSendMessage">
          <input 
            v-model="newMessageText" 
            type="text" 
            class="form-control" 
            placeholder="Scrivi un messaggio..."
            :disabled="isSending"
            autocomplete="off"
          >
          <button type="submit" class="btn btn-primary" :disabled="isSending">
            <LoadingSpinner v-if="isSending" />
            <svg v-else class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#send" /></svg>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>

/* Stili specifici per la Chat.vue */  

/* Layout della vista chat */
.chat-view {
  height: calc(100vh - 100px); 
}

/* Stili per l'intestazione della chat */
.chat-header {
  flex-shrink: 0;
}

/* Stili per l'area dei messaggi */
.message-list {
  flex-grow: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
}

/* Stili per i messaggi */
.message-wrapper {
  display: flex; 
  align-items: center;
  gap: 8px;
}

/* Allinea i messaggi inviati a destra */
.sent-wrapper {
  justify-content: flex-end;
  flex-direction: row-reverse; 
}

/* Stili per il pulsante di cancellazione del messaggio */
.delete-btn {
  border: none;
  opacity: 0; 
  transition: opacity 0.2s ease;
  padding: 4px;
}

/* Mostra il pulsante di cancellazione al passaggio del mouse */
.message-wrapper:hover .delete-btn {
  opacity: 1;
}

/* Stili per l'icona del cestino */
.delete-btn svg {
  width: 16px;
  height: 16px;
}

/* Stili per le bolle dei messaggi */
.message-bubble {
  background-color: #f1f0f0; 
  border-radius: 12px;
  padding: 10px 15px;
  margin-bottom: 10px;
  max-width: 70%;
  align-self: flex-start;
  word-wrap: break-word;
}

/* Stili per le bolle dei messaggi inviati */
.message-bubble.sent {
  background-color: #dcf8c6; 
  align-self: flex-end; 
}

/* Stili per il contenuto del messaggio */
.message-sender {
  font-size: 0.8rem;
  font-weight: bold;
  color: #075E54;
  margin-bottom: 4px;
}

/* Stili per il testo del messaggio */
.message-timestamp {
  font-size: 0.75rem;
  color: #999;
  text-align: right;
  margin-top: 5px;
}

/* Stili per l'area di input del messaggio */
.message-input-area {
  padding: 1rem;
  border-top: 1px solid #eee;
  flex-shrink: 0;
}

</style>