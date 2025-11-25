<script setup>
import { ref, computed } from 'vue';
// Aggiungi apiSetGroupPhoto agli import
import { apiSearchUsers, apiAddToGroup, apiLeaveGroup, apiSetGroupName, apiSetGroupPhoto } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const props = defineProps({
  show: Boolean,
  conversation: Object 
});

const emit = defineEmits(['close', 'refresh', 'left-group']);

// Stato Modifica Nome
const isEditingName = ref(false);
const newGroupName = ref('');
const loadingName = ref(false);

// Stato Modifica Foto (NUOVO)
const newPhotoUrl = ref('');
const loadingPhoto = ref(false);

// Stato Aggiunta Membri
const searchQuery = ref('');
const searchResults = ref([]);
const loadingSearch = ref(false);
const addingMemberId = ref(null);

const loadingLeave = ref(false);
const errorMsg = ref('');

// --- GESTIONE NOME GRUPPO ---
const startEditName = () => {
  newGroupName.value = props.conversation.name;
  isEditingName.value = true;
};

const saveGroupName = async () => {
  if (!newGroupName.value.trim()) return;
  loadingName.value = true;
  try {
    await apiSetGroupName(props.conversation.id, newGroupName.value);
    isEditingName.value = false;
    emit('refresh'); 
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loadingName.value = false;
  }
};

// --- GESTIONE FOTO GRUPPO (NUOVO) ---
const handleFileUpload = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  if (file.size > 1000000) { // Limite 1MB
    errorMsg.value = "L'immagine è troppo grande (max 1MB).";
    return;
  }

  const reader = new FileReader();
  reader.onload = async (e) => {
    const base64String = e.target.result;
    loadingPhoto.value = true;
    errorMsg.value = '';
    try {
      await apiSetGroupPhoto(props.conversation.id, base64String);
      emit('refresh'); // Ricarica per vedere la nuova foto
    } catch (err) {
      errorMsg.value = "Errore upload foto: " + err.message;
    } finally {
      loadingPhoto.value = false;
    }
  };
  reader.readAsDataURL(file);
};

// --- GESTIONE MEMBRI ---
const handleSearch = async () => {
  if (searchQuery.value.length < 1) return;
  loadingSearch.value = true;
  errorMsg.value = '';
  searchResults.value = [];
  
  try {
    const data = await apiSearchUsers(searchQuery.value);
    const currentMemberIds = new Set(props.conversation.members.map(m => m.id));
    searchResults.value = data.users.filter(u => !currentMemberIds.has(u.id)) || [];
    
    if (searchResults.value.length === 0) {
      errorMsg.value = "Nessun nuovo utente trovato.";
    }
  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loadingSearch.value = false;
  }
};

const handleAddMember = async (user) => {
  addingMemberId.value = user.id;
  errorMsg.value = '';
  try {
    await apiAddToGroup(props.conversation.id, user.id);
    searchQuery.value = '';
    searchResults.value = [];
    emit('refresh'); 
  } catch (err) {
    errorMsg.value = "Errore aggiunta: " + err.message;
  } finally {
    addingMemberId.value = null;
  }
};

const handleLeaveGroup = async () => {
  if (!confirm("Sei sicuro di voler abbandonare questo gruppo?")) return;
  loadingLeave.value = true;
  try {
    await apiLeaveGroup(props.conversation.id);
    emit('left-group'); 
  } catch (err) {
    errorMsg.value = err.message;
    loadingLeave.value = false;
  }
};
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="modal-title mb-0">Info Gruppo</h5>
        <button type="button" class="btn-close" @click="emit('close')" />
      </div>
      
      <div class="card-body">
        <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3"/>

        <div class="mb-4 text-center border-bottom pb-3">
            <div class="position-relative d-inline-block">
                <img 
                    :src="conversation.photoUrl || 'https://placehold.co/80x80/25d366/FFF?text=' + conversation.name.charAt(0)" 
                    class="rounded-circle border" width="80" height="80" style="object-fit: cover;"
                >
                <div v-if="loadingPhoto" class="position-absolute top-50 start-50 translate-middle">
                    <LoadingSpinner />
                </div>
                
                <label class="btn btn-sm btn-light position-absolute bottom-0 end-0 rounded-circle border shadow-sm p-1" style="cursor: pointer;" title="Cambia foto">
                    <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#camera" /></svg>
                    <input type="file" accept="image/*" class="d-none" @change="handleFileUpload">
                </label>
            </div>
        </div>

        <div class="mb-4 border-bottom pb-3">
          <label class="text-muted small fw-bold mb-2">NOME GRUPPO</label>
          
          <div v-if="!isEditingName" class="d-flex justify-content-between align-items-center">
            <h4 class="mb-0 text-truncate">{{ conversation.name }}</h4>
            <button class="btn btn-sm btn-outline-secondary" @click="startEditName">
              <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#edit-2" /></svg>
            </button>
          </div>

          <div v-else class="input-group">
            <input type="text" class="form-control" v-model="newGroupName" @keydown.enter.prevent="saveGroupName">
            <button class="btn btn-success" @click="saveGroupName" :disabled="loadingName">Salva</button>
            <button class="btn btn-outline-secondary" @click="isEditingName = false">Annulla</button>
          </div>
        </div>

        <div class="mb-4">
          <label class="text-muted small fw-bold mb-2">MEMBRI ({{ conversation.members.length }})</label>
          <div class="members-list border rounded p-2 bg-light">
            <div v-for="member in conversation.members" :key="member.id" class="d-flex align-items-center py-1">
              <img 
                :src="member.photoUrl || 'https://placehold.co/24x24/25d366/FFF?text=' + member.username.charAt(0)" 
                class="rounded-circle me-2" width="24" height="24"
              >
              <span>{{ member.username }}</span>
            </div>
          </div>
        </div>

        <div class="mb-4">
          <label class="text-muted small fw-bold mb-2">AGGIUNGI MEMBRO</label>
          <div class="input-group mb-2">
            <input 
              type="text" 
              class="form-control" 
              placeholder="Cerca utente..." 
              v-model="searchQuery"
              @keydown.enter.prevent="handleSearch"
            >
            <button class="btn btn-outline-primary" @click="handleSearch" :disabled="loadingSearch">
              <LoadingSpinner v-if="loadingSearch" />
              <svg v-else class="feather"><use href="/feather-sprite-v4.29.0.svg#search" /></svg>
            </button>
          </div>

          <div v-if="searchResults.length > 0" class="list-group search-results">
            <button 
              v-for="user in searchResults" 
              :key="user.id"
              class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
              @click="handleAddMember(user)"
              :disabled="addingMemberId === user.id"
            >
              <span>{{ user.username }}</span>
              <LoadingSpinner v-if="addingMemberId === user.id" />
              <span v-else class="badge bg-success text-white">+</span>
            </button>
          </div>
        </div>

        <div class="border-top pt-3">
          <button class="btn btn-outline-danger w-100" @click="handleLeaveGroup" :disabled="loadingLeave">
            <LoadingSpinner v-if="loadingLeave" />
            <span v-else>Abbandona Gruppo</span>
          </button>
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
  max-width: 450px;
  max-height: 90vh;
  overflow-y: auto;
  background-color: white; 
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
.members-list { max-height: 150px; overflow-y: auto; }
.search-results { max-height: 150px; overflow-y: auto; }
.feather { width: 16px; height: 16px; vertical-align: middle; }
</style>