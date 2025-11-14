<script setup>
import { ref, computed } from 'vue';
import { apiSearchUsers } from '@/services/api.js';
import { apiCreateGroup } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

// Props ed Eventi
defineProps({
  show: Boolean
});
const emit = defineEmits(['close', 'group-created']);

// Stato interno del modale
const groupName = ref('');
const searchQuery = ref('');
const searchResults = ref([]);
const selectedMembers = ref([]); // La lista dei membri da aggiungere
const loadingSearch = ref(false);
const loadingCreate = ref(false);
const errorMsg = ref('');

// Calcola se il form è valido per abilitare il pulsante "Crea"
const isFormValid = computed(() => {
  return groupName.value.length >= 1 && selectedMembers.value.length >= 1;
});

// Cerca utenti da aggiungere
const handleSearch = async () => {
  if (searchQuery.value.length < 1) return;
  
  loadingSearch.value = true;
  errorMsg.value = '';
  searchResults.value = [];
  
  try {
    const data = await apiSearchUsers(searchQuery.value);
    // Filtriamo gli utenti già selezionati
    const selectedIds = new Set(selectedMembers.value.map(m => m.id));
    searchResults.value = data.users.filter(user => !selectedIds.has(user.id)) || [];
    
    if (searchResults.value.length === 0) {
      errorMsg.value = "Nessun altro utente trovato.";
    }
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loadingSearch.value = false;
  }
};

// Aggiunge un utente dalla ricerca alla lista 'selectedMembers'
const addMember = (user) => {
  selectedMembers.value.push(user);
  searchResults.value = []; // Pulisce i risultati della ricerca
  searchQuery.value = ''; // Pulisce l'input di ricerca
};

// Rimuove un utente dalla lista 'selectedMembers'
const removeMember = (userId) => {
  selectedMembers.value = selectedMembers.value.filter(member => member.id !== userId);
};

// Funzione finale per creare il gruppo
const handleCreateGroup = async () => {
  if (!isFormValid.value) {
    errorMsg.value = "Inserisci un nome per il gruppo e almeno un membro.";
    return;
  }
  
  loadingCreate.value = true;
  errorMsg.value = '';

  try {
    // 1. Estrai solo gli ID dalla lista di oggetti 'selectedMembers'
    const memberIds = selectedMembers.value.map(member => member.id);
    
    // 2. Chiama l'API
    const newGroup = await apiCreateGroup(groupName.value, memberIds);
    
    // 3. Emetti l'evento di successo con l'ID del nuovo gruppo
    emit('group-created', newGroup.id);
    
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loadingCreate.value = false;
  }
};
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="modal-title mb-0">Crea un nuovo gruppo</h5>
        <button type="button" class="btn-close" @click="emit('close')" aria-label="Close"></button>
      </div>
      
      <div class="card-body">
        <form @submit.prevent="handleCreateGroup">
          
          <!-- Input Nome Gruppo -->
          <div class="mb-3">
            <label for="groupNameInput" class="form-label">Nome Gruppo:</label>
            <input 
              type="text" 
              class="form-control" 
              id="groupNameInput"
              placeholder="Es: Amici di WASA" 
              v-model="groupName"
              required>
          </div>

          <!-- Input Ricerca Membri -->
          <div class="mb-3">
            <label for="searchUserInput" class="form-label">Aggiungi Membri:</label>
            <div class="input-group">
              <input 
                type="text" 
                id="searchUserInput"
                class="form-control" 
                placeholder="Cerca un utente..." 
                v-model="searchQuery"
                @keydown.enter.prevent="handleSearch">
              <button class="btn btn-outline-secondary" type="button" @click="handleSearch" :disabled="loadingSearch">
                <LoadingSpinner v-if="loadingSearch" />
                <svg v-else class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
              </button>
            </div>
          </div>

          <!-- Risultati Ricerca -->
          <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
          <div v-if="searchResults.length > 0" class="list-group list-group-flush mb-3 search-results-box">
            <a 
              href="#"
              v-for="user in searchResults" 
              :key="user.id"
              @click.prevent="addMember(user)"
              class="list-group-item list-group-item-action d-flex gap-3 py-2 align-items-center">
              
              <img :src="user.photoUrl || 'https://placehold.co/40x40/25d366/FFF?text=' + user.username.charAt(0)" 
                   alt="foto" width="40" height="40" class="rounded-circle flex-shrink-0">
              <h6 class="mb-0">{{ user.username }}</h6>
            </a>
          </div>

          <!-- Membri Selezionati -->
          <div v-if="selectedMembers.length > 0" class="mb-3">
            <h6 class="text-muted small">MEMBRI ({{ selectedMembers.length }})</h6>
            <div class="selected-members-list">
              <span v-for="member in selectedMembers" :key="member.id" class="badge rounded-pill text-bg-primary me-2 mb-2">
                {{ member.username }}
                <button type="button" class="btn-close btn-close-white ms-1" @click="removeMember(member.id)" aria-label="Remove"></button>
              </span>
            </div>
          </div>

          <!-- Pulsante Crea Gruppo -->
          <button type="submit" class="btn btn-primary w-100" :disabled="!isFormValid || loadingCreate">
            <LoadingSpinner v-if="loadingCreate" />
            <span v-else>Crea Gruppo</span>
          </button>
          
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1050;
}

.modal-content {
  width: 100%;
  max-width: 500px;
  margin: 1rem;
  background-color: white; /* --- AGGIUNGI QUESTA RIGA --- */
  border-radius: 0.375rem; /* Aggiunge i bordi arrotondati del 'card' */
}

/* (il resto degli stili .search-results-box, .badge, ecc. rimane uguale) */
.search-results-box {
  max-height: 150px;
  overflow-y: auto;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
}
.selected-members-list {
  display: flex;
  flex-wrap: wrap;
}
.badge .btn-close {
  font-size: 0.65em;
  padding: 0.35em;
}
</style>