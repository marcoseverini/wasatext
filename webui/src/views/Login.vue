<template>
  <div class="d-flex justify-content-center align-items-center vh-100 bg-light">
    <div class="card p-4 shadow-sm" style="width: 100%; max-width: 400px;">
      <h1 class="h3 mb-3 fw-normal text-center">Benvenuto in WASAText</h1>
      <p class="text-center text-muted mb-4">Accedi o registrati per continuare</p>
      
      <form @submit.prevent="handleLogin">
        <div class="form-floating">
          <input 
            type="text" 
            class="form-control" 
            id="usernameInput" 
            placeholder="Il tuo nome utente"
            v-model="username"
            required>
          <label for="usernameInput">Nome utente</label>
        </div>

        <button 
          class="w-100 btn btn-lg btn-primary mt-3" 
          type="submit"
          :disabled="loading">
          <span v-if="loading" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
          <span v-else>Entra</span>
        </button>
        
        <p v-if="errorMsg" class="text-danger mt-3 text-center">{{ errorMsg }}</p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { apiLogin } from '@/services/api.js'; // Importa la nostra funzione API

const username = ref('');
const errorMsg = ref('');
const loading = ref(false);
const router = useRouter(); // Per il reindirizzamento

const handleLogin = async () => {
  loading.value = true;
  errorMsg.value = '';
  try {
    // Chiama la nostra funzione API centralizzata
    await apiLogin(username.value);
    
    // Login riuscito! Vai alla pagina Home
    router.push('/'); 
  } catch (err) {
    // Login fallito (l'errore arriva da api.js)
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
/* Puoi aggiungere stili specifici per il login qui */
</style>