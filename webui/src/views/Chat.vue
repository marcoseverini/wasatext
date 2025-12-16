<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'; 
import { useRoute, useRouter } from 'vue-router'; 
import { 
  apiGetConversation, 
  apiSendMessage,
  apiDeleteMessage,
  apiAddReaction,
  apiRemoveReaction
} from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue'; 
import LoadingSpinner from '@/components/LoadingSpinner.vue'; 
import GroupInfoModal from '@/components/GroupInfoModal.vue'; 
import ForwardModal from '@/components/ForwardModal.vue'; 

const conversation = ref(null); 
const loading = ref(true); 
const errorMsg = ref(''); 

// Invio
const newMessageText = ref(''); 
const selectedImageFile = ref(null);     
const selectedImagePreview = ref(null);  
const isSending = ref(false); 

const router = useRouter();
const showGroupInfo = ref(false); 

// Inoltro (AGGIORNATO: usa text e photo)
const showForwardModal = ref(false);
const msgTextToForward = ref(null);
const msgPhotoToForward = ref(null);

const replyingToMsg = ref(null); 
const availableEmojis = ['👍', '❤️', '😂', '😮', '😢', '🔥'];
const activeReactionMenuId = ref(null);

const route = useRoute(); 
const convId = route.params.id; 
const loggedInUserId = localStorage.getItem('sessionToken'); 

const messagesContainer = ref(null);
let pollingInterval = null;

const refreshConversation = async (showLoading = false) => {
  try {
    if (showLoading) loading.value = true;
    const data = await apiGetConversation(convId);
    if (data.messages) data.messages.reverse();
    
    const oldLastMsg = conversation.value?.messages?.at(-1)?.id;
    const newLastMsg = data.messages?.at(-1)?.id;
    
    conversation.value = data;

    if (oldLastMsg !== newLastMsg) scrollToBottom();
  } catch (err) {
    console.error("Refresh error:", err);
    if (showLoading) errorMsg.value = err.message;
  } finally {
    if (showLoading) loading.value = false;
  }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
};

onMounted(async () => { 
  await refreshConversation(true);
  scrollToBottom();
  pollingInterval = setInterval(() => refreshConversation(false), 3000);
});

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval);
});

const onFileSelect = (event) => {
  const file = event.target.files[0];
  if (!file) return;
  if (file.size > 1000000) { 
    alert("L'immagine è troppo grande (max 1MB).");
    return;
  }
  const reader = new FileReader();
  reader.onload = (e) => {
    selectedImagePreview.value = e.target.result; 
    selectedImageFile.value = file; 
  };
  reader.readAsDataURL(file);
  event.target.value = ''; 
};

const removeSelectedImage = () => {
  selectedImageFile.value = null;
  selectedImagePreview.value = null;
};

const handleSendMessage = async () => {
  if (newMessageText.value.trim() === '' && !selectedImageFile.value) return;

  isSending.value = true;
  errorMsg.value = '';
  const replyId = replyingToMsg.value ? replyingToMsg.value.id : null;

  try {
    // Chiamata API aggiornata (testo, foto)
    await apiSendMessage(
        convId, 
        newMessageText.value.trim(), 
        selectedImagePreview.value,  
        replyId
    );

    newMessageText.value = '';
    removeSelectedImage();
    replyingToMsg.value = null;
    await refreshConversation(false);
    scrollToBottom();

  } catch (err) {
    errorMsg.value = "Errore invio: " + err.message;
  } finally {
    isSending.value = false;
  }
};

const onLeftGroup = () => {
  showGroupInfo.value = false;
  router.push('/'); 
};

// Logica Inoltro Aggiornata
const openForwardModal = (msg) => {
  // Passiamo il contenuto esplicito al modale
  msgTextToForward.value = msg.text || '';
  msgPhotoToForward.value = msg.photoUrl || '';
  showForwardModal.value = true;
};

const onForwardSuccess = async () => {
  showForwardModal.value = false;
  msgTextToForward.value = null;
  msgPhotoToForward.value = null;
  await refreshConversation(false);
  scrollToBottom();
};

const handleDeleteMessage = async (messageId) => {
  if (!window.confirm("Cancellare messaggio?")) return;
  try {
    await apiDeleteMessage(messageId);
    await refreshConversation(false); 
  } catch (err) {
    errorMsg.value = err.message;
  }
};

const toggleReactionMenu = (msgId) => {
  activeReactionMenuId.value = activeReactionMenuId.value === msgId ? null : msgId;
};

const handleAddReaction = async (msgId, emoji) => {
  activeReactionMenuId.value = null; 
  try {
    await apiAddReaction(msgId, emoji);
    await refreshConversation(false); 
  } catch (err) {
    errorMsg.value = "Impossibile reagire: " + err.message;
  }
};

const handleRemoveReaction = async (msgId, reaction) => {
  if (reaction.user.id !== loggedInUserId) return;
  try {
    await apiRemoveReaction(msgId, reaction.id);
    await refreshConversation(false);
  } catch (err) {
    errorMsg.value = "Impossibile rimuovere: " + err.message;
  }
};

const startReply = (msg) => {
  replyingToMsg.value = msg;
};

const cancelReply = () => {
  replyingToMsg.value = null;
};

const getRepliedMessage = (replyId) => {
  return conversation.value.messages.find(m => m.id === replyId);
};
</script>

<template>
  <div class="chat-view">
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
    </div>

    <div v-if="!loading && conversation" class="d-flex flex-column h-100">
      
      <div class="d-flex align-items-center pt-3 pb-2 mb-3 border-bottom chat-header justify-content-between">
        <div class="d-flex align-items-center">
          <img
            :src="conversation.photoUrl || 'https://placehold.co/40x40/e9ecef/000000?text=' + conversation.name.charAt(0).toUpperCase()"
            alt="foto" width="40" height="40" class="rounded-circle me-3 border" style="object-fit: cover;"
          >
          <h1 class="h4 mb-0">{{ conversation.name }}</h1>
        </div>
        <button v-if="conversation.isGroup" class="btn btn-outline-secondary btn-sm info-btn" @click="showGroupInfo = true">
          <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#info" /></svg>
        </button>
      </div>

      <div class="message-list" @click="activeReactionMenuId = null" ref="messagesContainer"> 
        <div v-if="conversation.messages.length === 0" class="text-center text-muted">Inizio conversazione.</div>
        
        <div v-for="msg in conversation.messages" :key="msg.id"
          class="message-wrapper d-flex align-items-center"
          :class="{ 'sent-wrapper': msg.sender.id === loggedInUserId }"
        >
          <div class="message-bubble" :class="{ 'sent': msg.sender.id === loggedInUserId }"> 
            
            <div v-if="conversation.isGroup && msg.sender.id !== loggedInUserId" class="message-sender">
              {{ msg.sender.username }}
            </div>

            <div v-if="msg.replyToMsgId && getRepliedMessage(msg.replyToMsgId)" class="reply-preview-bubble mb-2">
              <div class="reply-line"></div>
              <div class="reply-content">
                <small class="fw-bold d-block text-primary">
                  {{ getRepliedMessage(msg.replyToMsgId).sender.username }}
                </small>
                <small class="text-truncate d-block" style="max-width: 200px;">
                  <span v-if="getRepliedMessage(msg.replyToMsgId).photoUrl">📷 [Foto] </span>
                  <span v-if="getRepliedMessage(msg.replyToMsgId).text">{{ getRepliedMessage(msg.replyToMsgId).text }}</span>
                </small>
              </div>
            </div>

            <div class="message-content">
              <div v-if="msg.photoUrl" class="mb-1">
                <img :src="msg.photoUrl" class="img-fluid rounded" style="max-width: 300px; max-height: 300px;">
              </div>
              <div v-if="msg.text" style="white-space: pre-wrap;">{{ msg.text }}</div>
            </div>
            
            <div v-if="msg.reactions && msg.reactions.length > 0" class="reactions-container mt-1">
              <span v-for="reaction in msg.reactions" :key="reaction.id"
                class="reaction-pill badge rounded-pill bg-light text-dark border"
                :class="{ 'my-reaction': reaction.user.id === loggedInUserId }"
                @click.stop="handleRemoveReaction(msg.id, reaction)"
              >
                {{ reaction.emoji }}
              </span>
            </div>

            <div class="message-footer d-flex align-items-center justify-content-end gap-1 mt-1">
              <small class="message-timestamp">
                {{ new Date(msg.timestamp).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' }) }}
              </small>
              <div v-if="msg.sender.id === loggedInUserId" class="message-status">
                <svg v-if="!msg.status" class="feather status-icon"><use href="/feather-sprite-v4.29.0.svg#clock" /></svg>
                <svg v-else-if="msg.status === 'sent' || msg.status === 'received'" class="feather status-icon"><use href="/feather-sprite-v4.29.0.svg#check" /></svg>
                <div v-else-if="msg.status === 'read'" class="d-flex">
                  <svg class="feather status-icon text-primary"><use href="/feather-sprite-v4.29.0.svg#check" /></svg>
                  <svg class="feather status-icon text-primary overlap-icon"><use href="/feather-sprite-v4.29.0.svg#check" /></svg>
                </div>
              </div>
            </div>
          </div>

          <div class="actions-group d-flex gap-1">
            <button class="btn btn-sm btn-outline-secondary action-btn" @click.stop="startReply(msg)">
              <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#corner-up-left" /></svg>
            </button>
            <button class="btn btn-sm btn-outline-secondary action-btn" @click.stop="openForwardModal(msg)">
              <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#share-2" /></svg>
            </button>
            <button v-if="msg.sender.id === loggedInUserId" class="btn btn-sm btn-outline-danger action-btn" @click.stop="handleDeleteMessage(msg.id)">
              <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2" /></svg>
            </button>
            <div class="position-relative">
              <button class="btn btn-sm btn-outline-secondary action-btn" @click.stop="toggleReactionMenu(msg.id)">
                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#smile" /></svg>
              </button>
              <div v-if="activeReactionMenuId === msg.id" class="emoji-picker shadow-sm">
                <span v-for="emoji in availableEmojis" :key="emoji" class="emoji-option" @click.stop="handleAddReaction(msg.id, emoji)">{{ emoji }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="message-input-area mt-auto">
        <div v-if="replyingToMsg" class="reply-bar alert alert-secondary d-flex justify-content-between align-items-center py-2 mb-2">
          <div class="d-flex align-items-center border-start border-4 border-primary ps-2">
            <div>
              <small class="fw-bold d-block text-primary">Rispondendo a {{ replyingToMsg.sender.username }}</small>
              <small class="text-muted text-truncate d-block" style="max-width: 300px;">
                 <span v-if="replyingToMsg.photoUrl">📷 [Foto] </span>
                 <span v-if="replyingToMsg.text">{{ replyingToMsg.text }}</span>
              </small>
            </div>
          </div>
          <button type="button" class="btn-close" @click="cancelReply"></button>
        </div>

        <div v-if="selectedImagePreview" class="p-2 mb-2 border rounded bg-light d-flex align-items-center justify-content-between">
          <div class="d-flex align-items-center gap-2">
            <img :src="selectedImagePreview" height="60" class="rounded border bg-white">
            <span class="small text-muted">Pronta per l'invio</span>
          </div>
          <button class="btn btn-sm btn-close" @click="removeSelectedImage"></button>
        </div>

        <form class="d-flex gap-2 align-items-center" @submit.prevent="handleSendMessage">
          <label class="btn btn-outline-secondary upload-btn">
            <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#camera" /></svg>
            <input type="file" accept="image/*" class="d-none" @change="onFileSelect" :disabled="isSending">
          </label>
          <input v-model="newMessageText" type="text" class="form-control" placeholder="Scrivi..." :disabled="isSending">
          <button type="submit" class="btn btn-primary" :disabled="isSending || (newMessageText.trim() === '' && !selectedImageFile)">
            <LoadingSpinner v-if="isSending" />
            <svg v-else class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#send" /></svg>
          </button>
        </form>
      </div>
    </div>

    <GroupInfoModal 
      v-if="showGroupInfo" :show="showGroupInfo" :conversation="conversation"
      @close="showGroupInfo = false" @refresh="refreshConversation" @left-group="onLeftGroup"
    />

    <ForwardModal 
      v-if="showForwardModal"
      :show="showForwardModal"
      :text="msgTextToForward"
      :photo="msgPhotoToForward"
      @close="showForwardModal = false"
      @forward-success="onForwardSuccess"
    />

  </div>
</template>

<style scoped>
/* (Stili identici al file precedente) */
.chat-view { height: calc(100vh - 100px); }
.chat-header { flex-shrink: 0; }
.message-list { flex-grow: 1; overflow-y: auto; padding: 1rem; display: flex; flex-direction: column; }
.message-wrapper { display: flex; align-items: flex-end; gap: 8px; margin-bottom: 10px; }
.sent-wrapper { flex-direction: row-reverse; }
.actions-group { opacity: 0; transition: opacity 0.2s ease; }
.message-wrapper:hover .actions-group, .active-menu .actions-group { opacity: 1; }
.action-btn, .info-btn { display: flex !important; align-items: center !important; justify-content: center !important; padding: 0 !important; width: 30px !important; height: 30px !important; border-radius: 4px; }
.action-btn svg, .info-btn svg { width: 16px; height: 16px; margin: 0 !important; vertical-align: middle; }
.upload-btn { display: flex !important; align-items: center !important; justify-content: center !important; padding: 0 !important; width: 38px !important; height: 38px !important; cursor: pointer; border-radius: 4px; }
.upload-btn svg { width: 20px; height: 20px; margin: 0 !important; vertical-align: middle; }
.emoji-picker { position: absolute; top: 35px; left: 0; background: white; border: 1px solid #ddd; border-radius: 8px; padding: 5px; display: flex; gap: 5px; z-index: 1000; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
.sent-wrapper .emoji-picker { left: auto; right: 0; }
.emoji-option { cursor: pointer; font-size: 1.2rem; padding: 2px 5px; border-radius: 4px; }
.emoji-option:hover { background-color: #f0f0f0; }
.message-bubble { background-color: #f1f0f0; border-radius: 12px; padding: 10px 15px; max-width: 70%; position: relative; word-wrap: break-word; }
.message-bubble.sent { background-color: #dcf8c6; }
.message-sender { font-size: 0.8rem; font-weight: bold; color: #075E54; margin-bottom: 4px; }
.reactions-container { display: flex; flex-wrap: wrap; gap: 4px; }
.reaction-pill { cursor: pointer; font-size: 0.85rem; padding: 2px 6px !important; border: 1px solid #ddd; }
.reaction-pill:hover { background-color: #e2e2e2 !important; }
.reaction-pill.my-reaction { background-color: #d1e7dd !important; border-color: #a3cfbb !important; }
.message-input-area { padding: 1rem; border-top: 1px solid #eee; flex-shrink: 0; }
.reply-bar { border-radius: 8px; font-size: 0.9rem; }
.reply-preview-bubble { background-color: rgba(0,0,0,0.05); border-radius: 6px; padding: 6px 10px; border-left: 4px solid #2470dc; font-size: 0.85rem; margin-bottom: 5px; }
.message-timestamp { font-size: 0.70rem; color: #999; }
.status-icon { width: 15px; height: 15px; color: #999; vertical-align: middle; margin: 0 !important; }
.text-primary { color: #0d6efd !important; }
.overlap-icon { margin-left: -8px !important; margin-top: -3px; }
</style>