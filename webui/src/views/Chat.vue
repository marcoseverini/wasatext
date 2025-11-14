<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { apiGetConversation } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Variabili reattive
const conversation = ref(null);
const loading = ref(true);
const errorMsg = ref('');

// useRoute() ci dà accesso ai parametri dell'URL
const route = useRoute();

// onMounted() viene eseguito quando la pagina carica
onMounted(async () => {
  try {
    loading.value = true;
    errorMsg.value = '';
    
    // 1. Leggi l'ID della chat dall'URL (es. /conversations/conv-123)
    const convId = route.params.id;
    
    // 2. Chiama l'API per ottenere i dettagli
    const data = await apiGetConversation(convId);
    conversation.value = data;
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="chat-view">
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <!-- Spinner di Caricamento -->
    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
    </div>

    <!-- Contenuto della Chat (quando caricato) -->
    <div v-if="!loading && !errorMsg && conversation">
      
      <!-- Intestazione Chat (Nome Gruppo o Utente) -->
      <div class="d-flex align-items-center pt-3 pb-2 mb-3 border-bottom">
        <img :src="conversation.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + conversation.name.charAt(0)"
             alt="foto" width="40" height="40" class="rounded-circle me-3">
        <h1 class="h4 mb-0">{{ conversation.name }}</h1>
      </div>

      <!-- Area Messaggi -->
      <div class="message-list">
        <div v-if="conversation.messages.length === 0" class="text-center text-muted">
          Questo è l'inizio della tua conversazione.
        </div>
        
        <!-- Itera sui messaggi (ricorda che il DB li dà dal più recente al meno recente) -->
        <div 
          v-for="msg in conversation.messages.slice().reverse()" 
          :key="msg.id"
          class="message-bubble"
          :class="{ 'sent': msg.sender.id === 'ID_UTENTE_LOGGATO' }"> 
          <!-- NB: Dobbiamo ancora implementare il controllo ID_UTENTE_LOGGATO -->
          
          <div class="message-sender" v-if="conversation.isGroup">{{ msg.sender.username }}</div>
          <div class="message-content">
            {{ msg.content }}
          </div>
          <div class="message-timestamp">
            {{ new Date(msg.timestamp).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' }) }}
          </div>
        </div>
      </div>
      
      <!-- Area Scrittura Messaggio (La faremo dopo) -->
      <div class="message-input-area">
        <input type="text" class="form-control" placeholder="Scrivi un messaggio...">
      </div>
      
    </div>
  </div>
</template>

<style scoped>
/* Stili semplici per una chat */
.chat-view {
  display: flex;
  flex-direction: column;
  height: 90vh; /* Altezza quasi totale */
}
.message-list {
  flex-grow: 1;
  overflow-y: auto; /* Permette lo scrolling */
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
  align-self: flex-start;
  word-wrap: break-word;
}
.message-bubble.sent {
  background-color: #dcf8c6; /* Messaggio inviato (verde) */
  align-self: flex-end;
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
}
</style>