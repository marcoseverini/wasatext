<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
// 1. Importa la nuova funzione
import { apiGetConversation, apiSendMessage } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Variabili reattive
const conversation = ref(null);
const loading = ref(true);
const errorMsg = ref('');

// 2. Aggiungi le variabili per il nuovo messaggio
const newMessageText = ref('');
const isSending = ref(false);

const route = useRoute();
const convId = route.params.id; // Prendiamo l'ID della chat dall'URL

// 3. Ci serve l'ID dell'utente loggato per lo stile
//    Il token salvato in localStorage È il nostro ID utente
const loggedInUserId = localStorage.getItem('sessionToken');

onMounted(async () => {
  try {
    loading.value = true;
    errorMsg.value = '';
    
    const data = await apiGetConversation(convId);
    
    // 4. Gira l'array dei messaggi
    // Il DB li dà (correttamente) dal più recente al meno recente.
    // Per mostrarli, li giriamo una volta sola al caricamento.
    if (data.messages) {
      data.messages.reverse();
    }
    conversation.value = data;
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
});

// 5. Funzione per inviare il messaggio
const handleSendMessage = async () => {
  if (newMessageText.value.trim() === '') return; // Non inviare messaggi vuoti
  
  isSending.value = true;
  errorMsg.value = '';
  
  try {
    // Chiama l'API
    const newMsg = await apiSendMessage(convId, newMessageText.value);
    
    // Aggiorna la UI in tempo reale!
    // Aggiungi il nuovo messaggio (appena creato)
    // in fondo alla lista dei messaggi.
    conversation.value.messages.push(newMsg);
    
    // Pulisci l'input
    newMessageText.value = '';

  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    isSending.value = false;
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
      
      <!-- Intestazione Chat (Nome Gruppo o Utente) -->
      <div class="d-flex align-items-center pt-3 pb-2 mb-3 border-bottom chat-header">
        <img :src="conversation.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + conversation.name.charAt(0)"
             alt="foto" width="40" height="40" class="rounded-circle me-3">
        <h1 class="h4 mb-0">{{ conversation.name }}</h1>
      </div>

      <!-- Area Messaggi -->
      <div class="message-list">
        <div v-if="conversation.messages.length === 0" class="text-center text-muted">
          Questo è l'inizio della tua conversazione.
        </div>
        
        <!-- 6. Itera sui messaggi (ora non serve .reverse()) -->
        <div 
          v-for="msg in conversation.messages" 
          :key="msg.id"
          class="message-bubble"
          :class="{ 'sent': msg.sender.id === loggedInUserId }"> 
          <!-- Applica la classe 'sent' se il mittente siamo noi -->
          
          <div class="message-sender" v-if="conversation.isGroup && msg.sender.id !== loggedInUserId">
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
      
      <!-- 7. Area Scrittura Messaggio (ora collegata) -->
      <div class="message-input-area mt-auto">
        <form @submit.prevent="handleSendMessage" class="d-flex gap-2">
          <input 
            type="text" 
            class="form-control" 
            placeholder="Scrivi un messaggio..." 
            v-model="newMessageText"
            :disabled="isSending"
            autocomplete="off">
          <button type="submit" class="btn btn-primary" :disabled="isSending">
            <LoadingSpinner v-if="isSending" />
            <!-- Icona Invia (feather) -->
            <svg v-else class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#send"/></svg>
          </button>
        </form>
      </div>
      
    </div>
  </div>
</template>

<style scoped>
.chat-view {
  /* Fai in modo che il contenitore prenda tutta l'altezza del 'main' */
  height: calc(100vh - 100px); /* 100vh meno l'header e il padding */
}
.chat-header {
  flex-shrink: 0;
}
.message-list {
  flex-grow: 1; /* Occupa tutto lo spazio disponibile */
  overflow-y: auto; /* Aggiunge lo scrolling solo ai messaggi */
  padding: 1rem;
  display: flex;
  flex-direction: column;
}
.message-bubble {
  background-color: #f1f0f0; /* Messaggio ricevuto (grigio) */
  border-radius: 12px;
  padding: 10px 15px;
  margin-bottom: 10px;
  max-width: 70%;
  align-self: flex-start; /* Default: allineato a sinistra (ricevuto) */
  word-wrap: break-word;
}
.message-bubble.sent {
  background-color: #dcf8c6; /* Messaggio inviato (verde) */
  align-self: flex-end; /* Allineato a destra (inviato) */
}
.message-sender {
  font-size: 0.8rem;
  font-weight: bold;
  color: #075E54;
  margin-bottom: 4px;
}
.message-timestamp {
  font-size: 0.75rem;
  color: #999;
  text-align: right;
  margin-top: 5px;
}
.message-input-area {
  padding: 1rem;
  border-top: 1px solid #eee;
  flex-shrink: 0; /* Non ridurre questa barra */
}
</style>