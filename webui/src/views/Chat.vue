<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
// 1. Importa la nuova funzione
import { apiGetConversation, apiSendMessage, apiDeleteMessage } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const conversation = ref(null);
const loading = ref(true);
const errorMsg = ref('');
const newMessageText = ref('');
const isSending = ref(false);

const route = useRoute();
const convId = route.params.id;
const loggedInUserId = localStorage.getItem('sessionToken');

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

// 2. Aggiungi la funzione per cancellare il messaggio
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
      
      <!-- Intestazione Chat (invariata) -->
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
        
        <!-- 3. Modifica il template per includere il pulsante Cestino -->
        <div 
          v-for="msg in conversation.messages" 
          :key="msg.id"
          class="message-wrapper d-flex align-items-center"
          :class="{ 'sent-wrapper': msg.sender.id === loggedInUserId }"> 
          
          <!-- Pulsante Cestino (mostrato solo se 'sent') -->
          <button 
            v-if="msg.sender.id === loggedInUserId"
            @click="handleDeleteMessage(msg.id)"
            class="btn btn-sm btn-outline-danger delete-btn">
            <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
          </button>
          
          <!-- Bolla del Messaggio (invariata) -->
          <div class="message-bubble" :class="{ 'sent': msg.sender.id === loggedInUserId }"> 
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
      </div>
      
      <!-- Area Scrittura Messaggio (invariata) -->
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
            <svg v-else class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#send"/></svg>
          </button>
        </form>
      </div>
      
    </div>
  </div>
</template>

<style scoped>
.chat-view {
  height: calc(100vh - 100px); 
}
.chat-header {
  flex-shrink: 0;
}
.message-list {
  flex-grow: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
}

/* --- 4. Aggiunti stili per il pulsante Cestino --- */
.message-wrapper {
  gap: 8px;
}
/* Allinea a destra i messaggi inviati (bolla + pulsante) */
.sent-wrapper {
  justify-content: flex-end;
  flex-direction: row-reverse; /* Inverte l'ordine: prima la bolla, poi il pulsante */
}
.delete-btn {
  border: none;
  opacity: 0.1; /* Nascosto di default */
  transition: opacity 0.2s ease;
  padding: 4px;
}
.message-wrapper:hover .delete-btn {
  opacity: 1; /* Appare in hover */
}
.delete-btn svg {
  width: 16px;
  height: 16px;
}
/* --- Fine stili Cestino --- */


.message-bubble {
  background-color: #f1f0f0; 
  border-radius: 12px;
  padding: 10px 15px;
  margin-bottom: 10px;
  max-width: 70%;
  align-self: flex-start;
  word-wrap: break-word;
}
.message-bubble.sent {
  background-color: #dcf8c6; 
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
  flex-shrink: 0;
}
</style>