<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
// 1. Importa le nuove funzioni
import { 
  apiGetConversation, 
  apiSendMessage, 
  apiCommentMessage, 
  apiUncommentMessage,
  apiDeleteMessage
} from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Variabili reattive
const conversation = ref(null);
const loading = ref(true);
const errorMsg = ref('');
const newMessageText = ref('');
const isSending = ref(false);

const route = useRoute();
const convId = route.params.id; 
const loggedInUserId = localStorage.getItem('sessionToken');

// 2. Nuove variabili per le reazioni
const activePickerMsgId = ref(null); // Traccia quale selettore emoji è aperto
const simpleEmojiPicker = ['👍', '❤️', '😂', '😮', '😢', '🙏']; // Il nostro selettore

// Funzione per caricare la chat (invariata)
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

// Funzione per inviare il messaggio (invariata)
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

// Funzione per cancellare il messaggio (invariata)
const handleDeleteMessage = async (messageId) => {
  if (!window.confirm("Sei sicuro di voler cancellare questo messaggio?")) {
    return;
  }
  errorMsg.value = '';
  try {
    await apiDeleteMessage(messageId);
    conversation.value.messages = conversation.value.messages.filter(
      (msg) => msg.id !== messageId
    );
  } catch (err) {
    errorMsg.value = err.message;
  }
};


// --- 3. NUOVE FUNZIONI PER LE REAZIONI ---

/**
 * Apre/Chiude il selettore emoji per un messaggio
 */
const togglePicker = (messageId) => {
  if (activePickerMsgId.value === messageId) {
    activePickerMsgId.value = null; // Chiude se è già aperto
  } else {
    activePickerMsgId.value = messageId; // Apre
  }
};

/**
 * Chiamato quando l'utente clicca su un emoji nel selettore
 */
const handleCommentMessage = async (messageId, emoji) => {
  activePickerMsgId.value = null; // Chiude il selettore
  errorMsg.value = '';

  try {
    // Chiama l'API per aggiungere la reazione
    const newReaction = await apiCommentMessage(messageId, emoji);
    
    // Aggiorna la UI in tempo reale:
    // 1. Trova il messaggio nella nostra lista
    const msg = conversation.value.messages.find(m => m.id === messageId);
    if (msg) {
      // 2. Aggiungi la nuova reazione al suo array (se non esiste già per ID)
      if (!msg.reactions.find(r => r.id === newReaction.id)) {
        msg.reactions.push(newReaction);
      }
    }
  } catch (err) {
    errorMsg.value = err.message;
  }
};

/**
 * Chiamato quando l'utente clicca su una reazione ESISTENTE
 */
const handleUncommentMessage = async (message, reaction) => {
  // Permetti la cancellazione SOLO se l'utente è il proprietario della reazione
  if (reaction.user.id !== loggedInUserId) {
    return; 
  }

  errorMsg.value = '';
  try {
    // Chiama l'API
    await apiUncommentMessage(message.id, reaction.id);
    
    // Aggiorna la UI in tempo reale:
    // Rimuovi la reazione dall'array
    message.reactions = message.reactions.filter(r => r.id !== reaction.id);

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
        
        <!-- Itera sui messaggi -->
        <div 
          v-for="msg in conversation.messages" 
          :key="msg.id"
          class="message-wrapper"
          :class="{ 'sent-wrapper': msg.sender.id === loggedInUserId }"> 
          
          <!-- Pulsante Cestino (come prima) -->
          <button 
            v-if="msg.sender.id === loggedInUserId"
            @click="handleDeleteMessage(msg.id)"
            class="btn btn-sm btn-outline-danger delete-btn">
            <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
          </button>
          
          <!-- --- ▼▼▼ 4. NUOVO BLOCCO REAZIONI ▼▼▼ --- -->
          <!-- Pulsante "Aggiungi Reazione" (+) -->
          <button
            @click="togglePicker(msg.id)"
            class="btn btn-sm btn-outline-secondary react-btn">
            <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#smile"/></svg>
          </button>
          
          <!-- Selettore Emoji (appare quando 'activePickerMsgId' corrisponde) -->
          <div v-if="activePickerMsgId === msg.id" class="emoji-picker">
            <span 
              v-for="emoji in simpleEmojiPicker" 
              :key="emoji" 
              @click="handleCommentMessage(msg.id, emoji)">
              {{ emoji }}
            </span>
          </div>
          <!-- --- ▲▲▲ FINE BLOCCO REAZIONI ▲▲▲ --- -->


          <!-- Bolla del Messaggio (come prima) -->
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
            
            <!-- --- ▼▼▼ 5. NUOVO BLOCCO MOSTRA REAZIONI ▼▼▼ --- -->
            <div v-if="msg.reactions.length > 0" class="reactions-list">
              <span 
                v-for="reaction in msg.reactions" 
                :key="reaction.id"
                class="reaction-badge"
                :class="{ 'my-reaction': reaction.user.id === loggedInUserId }"
                @click="handleUncommentMessage(msg, reaction)">
                {{ reaction.emoji }}
                <span class="reaction-count">{{ msg.reactions.filter(r => r.emoji === reaction.emoji).length }}</span>
              </span>
            </div>
            <!-- --- ▲▲▲ FINE BLOCCO MOSTRA REAZIONI ▲▲▲ --- -->
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
/* (Stili per .chat-view, .chat-header, .message-list, .message-input-area... rimangono invariati) */
.chat-view { height: calc(100vh - 100px); }
.chat-header { flex-shrink: 0; }
.message-list { flex-grow: 1; overflow-y: auto; padding: 1rem; display: flex; flex-direction: column; }
.message-input-area { padding: 1rem; border-top: 1px solid #eee; flex-shrink: 0; }

/* Stili Wrapper Messaggio */
.message-wrapper {
  display: flex; /* Cambiato da 'gap' a 'display:flex' per un controllo migliore */
  align-items: center;
  position: relative; /* Necessario per posizionare il selettore emoji */
}
.sent-wrapper {
  justify-content: flex-end;
  flex-direction: row-reverse; 
}

/* Stili Pulsanti Azione (Delete, React) */
.delete-btn, .react-btn {
  border: none;
  opacity: 0; /* Nascosti di default */
  transition: opacity 0.2s ease;
  padding: 4px;
  background: #fff;
  border-radius: 50%;
  margin: 0 4px;
}
.message-wrapper:hover .delete-btn,
.message-wrapper:hover .react-btn {
  opacity: 1; /* Appaiono in hover */
}
.delete-btn svg, .react-btn svg {
  width: 16px;
  height: 16px;
}

/* Bolla Messaggio */
.message-bubble {
  background-color: #f1f0f0; 
  border-radius: 12px;
  padding: 10px 15px;
  margin-bottom: 10px;
  max-width: 70%;
  align-self: flex-start;
  word-wrap: break-word;
  position: relative; /* Necessario per le reazioni */
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

/* --- 6. NUOVI STILI PER LE REAZIONI --- */
.emoji-picker {
  position: absolute;
  bottom: 100%; /* Appare sopra la bolla */
  left: 40px; /* Posizionato vicino al pulsante + */
  background: white;
  border: 1px solid #ccc;
  border-radius: 8px;
  padding: 8px;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
  display: flex;
  gap: 8px;
  z-index: 10;
}
.sent-wrapper .emoji-picker {
  left: auto; /* Se inviato, allinea a destra */
  right: 40px; 
}
.emoji-picker span {
  font-size: 1.5rem;
  cursor: pointer;
  transition: transform 0.1s ease;
}
.emoji-picker span:hover {
  transform: scale(1.2);
}

.reactions-list {
  position: absolute;
  bottom: -15px; /* Sovrappone leggermente la bolla successiva */
  left: 10px;
  display: flex;
  gap: 4px;
}
.sent .reactions-list {
  left: auto;
  right: 10px;
}
.reaction-badge {
  background-color: #fff;
  border: 1px solid #ddd;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}
.reaction-badge.my-reaction {
  background-color: #e0f2ff; /* Evidenzia le mie reazioni */
  border-color: #007bff;
  cursor: pointer; /* Indica che posso cancellarla */
}
.reaction-count {
  font-size: 0.7rem;
  font-weight: bold;
  margin-left: 4px;
  color: #333;
}
</style>