<template> 
  
  <div class="d-flex justify-content-center align-items-center vh-100 bg-light"> <!-- Centra il contenuto verticalmente e orizzontalmente -->
    <div class="card p-4 shadow-sm" style="width: 100%; max-width: 400px;"> <!-- Card per il modulo di login -->
      <h1 class="h3 mb-3 fw-normal text-center">Benvenuto in WASAText</h1> <!-- Titolo della pagina -->
      <p class="text-center text-muted mb-4">Accedi o registrati per continuare</p> <!-- Sottotitolo -->
      
       <!-- Gestisce il submit del modulo -->
      <form @submit.prevent="handleLogin">
        <div class="form-floating">  <!-- Campo di input per il nome utente -->
          <input 
            id="usernameInput" 
            v-model="username" 
            type="text" 
            class="form-control"
            placeholder="Il tuo nome utente"
            required
          >
          <label for="usernameInput">Nome utente</label> <!-- Etichetta del campo -->
        </div>

        <!-- Pulsante di invio del modulo -->
        <button 
          class="w-100 btn btn-lg btn-primary mt-3" 
          type="submit"
          :disabled="loading"
        > 
          <span v-if="loading" class="spinner-border spinner-border-sm" role="status" aria-hidden="true" />
          <span v-else>Entra</span>
        </button> 
        
        <!-- Messaggio di errore -->
        <p v-if="errorMsg" class="text-danger mt-3 text-center">{{ errorMsg }}</p>

      </form>
    </div>
  </div>
</template>

<script setup>

import { ref } from 'vue'; // Per le variabili reattive
import { useRouter } from 'vue-router'; // Per il routing
import { apiLogin } from '@/services/api.js'; // Importa la nostra funzione API

const username = ref(''); // Variabile reattiva per il nome utente
const errorMsg = ref(''); // Variabile reattiva per il messaggio di errore
const loading = ref(false); // Variabile reattiva per lo stato di caricamento
const router = useRouter(); // Ottiene l'istanza del router

const handleLogin = async () => { // Funzione per gestire il login
  loading.value = true; // Imposta lo stato di caricamento a true
  errorMsg.value = ''; // Resetta il messaggio di errore
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
/* Potrei aggiungere stili specifici per il login qui */
</style>