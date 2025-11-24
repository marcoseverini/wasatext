<script setup>
import { ref, onMounted } from 'vue'; 
import { useRoute } from 'vue-router'; 
import { 
  apiGetConversation, 
  apiSendMessage,
  apiDeleteMessage,
  apiAddReaction,
  apiRemoveReaction
} from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue'; 
import LoadingSpinner from '@/components/LoadingSpinner.vue'; 

const conversation = ref(null); 
const loading = ref(true); 
const errorMsg = ref(''); 
const newMessageText = ref(''); 
const isSending = ref(false); 

const availableEmojis = ['👍', '❤️', '😂', '😮', '😢', '🔥'];
const activeReactionMenuId = ref(null);

const route = useRoute(); 
const convId = route.params.id; 
const loggedInUserId = localStorage.getItem('sessionToken'); 

onMounted(async () => { 
  try {
    loading.value = true;
    errorMsg.value = '';
    const data = await apiGetConversation(convId);
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

const handleDeleteMessage = async (messageId) => {
  if (!window.confirm("Sei sicuro di voler cancellare questo messaggio?")) return;
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

const toggleReactionMenu = (msgId) => {
  if (activeReactionMenuId.value === msgId) {
    activeReactionMenuId.value = null;
  } else {
    activeReactionMenuId.value = msgId;
  }
};

const handleAddReaction = async (msgId, emoji) => {
  activeReactionMenuId.value = null; 
  try {
    const reaction = await apiAddReaction(msgId, emoji);
    const msg = conversation.value.messages.find(m => m.id === msgId);
    if (msg) {
      if (!msg.reactions) msg.reactions = [];
      msg.reactions = msg.reactions.filter(r => r.user.id !== loggedInUserId);
      msg.reactions.push(reaction);
    }
  } catch (err) {
    errorMsg.value = "Impossibile reagire: " + err.message;
  }
};

const handleRemoveReaction = async (msgId, reaction) => {
  if (reaction.user.id !== loggedInUserId) return;
  try {
    await apiRemoveReaction(msgId, reaction.id);
    const msg = conversation.value.messages.find(m => m.id === msgId);
    if (msg && msg.reactions) {
      msg.reactions = msg.reactions.filter(r => r.id !== reaction.id);
    }
  } catch (err) {
    errorMsg.value = "Impossibile rimuovere reazione: " + err.message;
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
      
      <div class="d-flex align-items-center pt-3 pb-2 mb-3 border-bottom chat-header">
        <img
          :src="conversation.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + conversation.name.charAt(0)"
          alt="foto" width="40" height="40" class="rounded-circle me-3"
        >
        <h1 class="h4 mb-0">{{ conversation.name }}</h1>
      </div>

      <div class="message-list" @click="activeReactionMenuId = null"> 
        <div v-if="conversation.messages.length === 0" class="text-center text-muted">
          Questo è l'inizio della tua conversazione.
        </div>
        
        <div 
          v-for="msg in conversation.messages" 
          :key="msg.id"
          class="message-wrapper d-flex align-items-center"
          :class="{ 'sent-wrapper': msg.sender.id === loggedInUserId }"
        >
          
          <div class="message-bubble" :class="{ 'sent': msg.sender.id === loggedInUserId }"> 
            <div v-if="conversation.isGroup && msg.sender.id !== loggedInUserId" class="message-sender">
              {{ msg.sender.username }}
            </div>
            <div class="message-content">
              {{ msg.content }}
            </div>
            
            <div v-if="msg.reactions && msg.reactions.length > 0" class="reactions-container mt-1">
              <span 
                v-for="reaction in msg.reactions" 
                :key="reaction.id"
                class="reaction-pill badge rounded-pill bg-light text-dark border"
                :class="{ 'my-reaction': reaction.user.id === loggedInUserId }"
                @click.stop="handleRemoveReaction(msg.id, reaction)"
                :title="reaction.user.username"
              >
                {{ reaction.emoji }}
              </span>
            </div>

            <div class="message-timestamp">
              {{ new Date(msg.timestamp).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' }) }}
            </div>
          </div>

          <div class="actions-group d-flex gap-1">
            
            <button 
              v-if="msg.sender.id === loggedInUserId"
              class="btn btn-sm btn-outline-danger action-btn"
              @click.stop="handleDeleteMessage(msg.id)"
            >
              <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2" /></svg>
            </button>

            <div class="position-relative">
              <button 
                class="btn btn-sm btn-outline-secondary action-btn"
                @click.stop="toggleReactionMenu(msg.id)"
              >
                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#smile" /></svg>
              </button>

              <div v-if="activeReactionMenuId === msg.id" class="emoji-picker shadow-sm">
                <span 
                  v-for="emoji in availableEmojis" 
                  :key="emoji"
                  class="emoji-option"
                  @click.stop="handleAddReaction(msg.id, emoji)"
                >
                  {{ emoji }}
                </span>
              </div>
            </div>

          </div>

        </div>
      </div>
      
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

/* Wrapper dei messaggi */
.message-wrapper {
  display: flex; 
  align-items: flex-end; 
  /* Usiamo gap per separare bolla e bottoni indipendentemente dalla direzione */
  gap: 8px; 
  margin-bottom: 10px;
}

/* --- LOGICA DI ALLINEAMENTO --- */

/* Messaggi INVIATI (Miei) */
.sent-wrapper {
  /* row-reverse fa due cose magiche qui:
     1. Inverte l'ordine visivo: [Bottoni] [Bolla] (i bottoni vanno a sinistra della bolla)
     2. Inverte l'asse principale: "Start" diventa Destra. Quindi si allineano a destra.
  */
  flex-direction: row-reverse;
}
/* Messaggi RICEVUTI (Altri) */
/* Di default è 'row', quindi: [Bolla] [Bottoni]. Start è Sinistra. Perfetto così. */


/* Gruppo bottoni */
.actions-group {
  opacity: 0; 
  transition: opacity 0.2s ease;
}
.message-wrapper:hover .actions-group, 
.active-menu .actions-group { 
  opacity: 1;
}

.action-btn {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  padding: 0 !important;
  width: 30px !important;
  height: 30px !important;
  border-radius: 4px;
}
.action-btn svg {
  width: 16px;
  height: 16px;
  /* --- FIX CENTRAGGIO --- */
  margin: 0 !important;       /* Rimuove il margine destro ereditato da App.vue */
  vertical-align: middle;     /* Assicura l'allineamento verticale preciso */
  /* ---------------------- */
}

/* MENU EMOJI POPUP */
.emoji-picker {
  position: absolute;
  top: 35px;
  
  /* DEFAULT: Allineato a SINISTRA (Va bene per i messaggi ricevuti/altri) */
  left: 0; 
  
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 5px;
  display: flex;
  gap: 5px;
  z-index: 1000;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

/* --- NUOVA REGOLA: OVERRIDE PER MESSAGGI INVIATI --- */
/* Se il menu è dentro un messaggio inviato (.sent-wrapper),
   annulliamo 'left' e usiamo 'right' per farlo crescere verso sinistra/centro. */
.sent-wrapper .emoji-picker {
  left: auto;  /* Disabilita il left: 0 di default */
  right: 0;    /* Allinea il bordo destro del menu al bordo destro del bottone */
}

.emoji-option {
  cursor: pointer;
  font-size: 1.2rem;
  padding: 2px 5px;
  border-radius: 4px;
}
.emoji-option:hover {
  background-color: #f0f0f0;
}

.message-bubble {
  background-color: #f1f0f0; 
  border-radius: 12px;
  padding: 10px 15px;
  max-width: 70%;
  position: relative;
  word-wrap: break-word;
}
.message-bubble.sent {
  background-color: #dcf8c6; 
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

.reactions-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.reaction-pill {
  cursor: pointer;
  font-size: 0.85rem;
  padding: 2px 6px !important;
  border: 1px solid #ddd;
}
.reaction-pill:hover {
  background-color: #e2e2e2 !important;
}
.reaction-pill.my-reaction {
  background-color: #d1e7dd !important; 
  border-color: #a3cfbb !important;
}

.message-input-area {
  padding: 1rem;
  border-top: 1px solid #eee;
  flex-shrink: 0;
}
</style>