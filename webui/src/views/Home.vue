<template>
  <div>
    <!-- Questo è il titolo che vedi nel layout principale -->
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Le mie Conversazioni</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <!-- Qui potremmo mettere un pulsante "Cerca Utenti" o "Crea Gruppo" -->
      </div>
    </div>

    <!-- Contenuto della pagina -->
    <div v-if="loading" class="text-center">
      <LoadingSpinner />
    </div>
    
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    
    <div v-if="!loading && !errorMsg">
      <!-- Qui è dove caricheremo e mostreremo la lista delle chat -->
      <p>Caricamento delle conversazioni...</p>
      <pre>{{ conversations }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { apiGetMyConversations } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const conversations = ref(null);
const loading = ref(true);
const errorMsg = ref('');

// onMounted() è l'equivalente di "quando la pagina ha caricato"
onMounted(async () => {
  try {
    loading.value = true;
    errorMsg.value = '';
    
    // Chiama l'API per ottenere le conversazioni
    const data = await apiGetMyConversations();
    conversations.value = data.conversations; // Basato sul refactoring dello YAML
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
});
</script>