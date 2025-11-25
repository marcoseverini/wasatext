<script setup>
import { ref, onMounted } from 'vue';
import { apiGetMyConversations, apiForwardMessage } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const props = defineProps({
  show: Boolean,
  messageId: String // L'ID del messaggio da inoltrare
});

const emit = defineEmits(['close', 'forward-success']);

const conversations = ref([]);
const loading = ref(true);
const sending = ref(false);
const errorMsg = ref('');

// Carica la lista delle chat dove posso inoltrare
onMounted(async () => {
  try {
    loading.value = true;
    const data = await apiGetMyConversations();
    conversations.value = data.conversations || [];
  } catch (err) {
    errorMsg.value = "Impossibile caricare le conversazioni: " + err.message;
  } finally {
    loading.value = false;
  }
});

const handleForward = async (targetConvId) => {
  if (sending.value) return;
  sending.value = true;
  
  try {
    await apiForwardMessage(targetConvId, props.messageId);
    alert("Messaggio inoltrato!");
    emit('forward-success'); // Chiude il modale
  } catch (err) {
    alert("Errore inoltro: " + err.message);
  } finally {
    sending.value = false;
  }
};
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="modal-title mb-0">Inoltra a...</h5>
        <button type="button" class="btn-close" @click="emit('close')" />
      </div>
      
      <div class="card-body p-0"> <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="m-3"/>

        <div v-if="loading" class="text-center p-4">
          <LoadingSpinner />
        </div>

        <div v-else class="list-group list-group-flush">
          <button 
            v-for="convo in conversations" 
            :key="convo.id"
            class="list-group-item list-group-item-action d-flex align-items-center gap-3 py-3"
            @click="handleForward(convo.id)"
            :disabled="sending"
          >
            <img 
              :src="convo.photoUrl || 'https://placehold.co/40x40/e9ecef/000000?text=' + convo.name.charAt(0).toUpperCase()" 
              class="rounded-circle flex-shrink-0 border" width="40" height="40"
              style="object-fit: cover;"
            >
            <div class="text-truncate fw-bold">{{ convo.name }}</div>
            
            </button>

          <div v-if="conversations.length === 0" class="text-center p-4 text-muted">
            Nessuna conversazione attiva.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0;
  width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex; justify-content: center; align-items: center;
  z-index: 2000;
}
.modal-content {
  max-width: 400px;
  max-height: 80vh;
  overflow-y: auto;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
</style>