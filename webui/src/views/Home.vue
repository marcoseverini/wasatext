<script setup>
import { ref, onMounted } from 'vue'; 
import { useRouter } from 'vue-router'; 
import { apiGetMyConversations } from '@/services/api.js'; 
import ErrorMsg from '@/components/ErrorMsg.vue'; 
import LoadingSpinner from '@/components/LoadingSpinner.vue'; 

const conversations = ref([]); 
const loading = ref(true); 
const errorMsg = ref(''); 
const router = useRouter(); 

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

const goToConversation = (convId) => {
  router.push(`/conversations/${convId}`);
};

const formatTimestamp = (isoString) => {
  if (!isoString) return '';
  const date = new window.Date(isoString);
  const today = new window.Date();
  const isToday = date.getDate() === today.getDate() &&
                  date.getMonth() === today.getMonth() &&
                  date.getFullYear() === today.getFullYear();
  if (isToday) {
    return date.toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' });
  } else {
    return date.toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit' });
  }
};

const formatSnippet = (snippet) => {
  if (!snippet) return 'Nessun messaggio';
  if (snippet.startsWith('data:image')) return '📷 [Foto]';
  return snippet;
};

onMounted(() => {
  loadConversations();
});
</script>

<template>
  <div>
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Conversazioni</h1>
    </div>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />

    <div v-if="loading" class="text-center mt-5">
      <LoadingSpinner />
      <p class="text-muted mt-2">Caricamento...</p>
    </div>

    <div v-if="!loading && !errorMsg">
      <div v-if="conversations.length === 0" class="text-center text-muted mt-5 p-5">
        <h4>Nessuna conversazione</h4>
        <p>Usa il menu a sinistra per cercare utenti o creare un gruppo!</p>
      </div>

      <div v-else class="list-group shadow-sm">
        <a 
          v-for="convo in conversations"
          :key="convo.id" 
          href="#"
          class="list-group-item list-group-item-action d-flex gap-3 py-3 align-items-center"
          @click.prevent="goToConversation(convo.id)"
        >
          <img
            :src="convo.photoUrl || 'https://placehold.co/64x64/e9ecef/000000?text=' + convo.name.charAt(0).toUpperCase()" 
            alt="foto" width="50" height="50" class="rounded-circle flex-shrink-0 border"
            style="object-fit: cover;"
          >
          
          <div class="d-flex gap-2 w-100 justify-content-between overflow-hidden">
            <div class="overflow-hidden">
              <h6 class="mb-0 text-truncate">{{ convo.name }}</h6>
              <p class="mb-0 opacity-75 text-truncate small text-muted">
                {{ formatSnippet(convo.latestMessageSnippet) }}
              </p>
            </div>
            <small class="opacity-50 text-nowrap" style="font-size: 0.8rem;">
              {{ formatTimestamp(convo.latestMessageTimestamp) }}
            </small>
          </div>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>