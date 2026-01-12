<script setup>
import { ref, watch } from 'vue';
import { apiSetMyUserName, apiSetMyPhoto } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const props = defineProps({
  show: Boolean,
  username: String,
  photoUrl: String 
});

const emit = defineEmits(['close', 'update-profile']);

const newUsername = ref(props.username);
const newPhotoUrl = ref(props.photoUrl || ''); 
const loading = ref(false);
const errorMsg = ref('');
const successMsg = ref('');

watch(() => props.username, (val) => newUsername.value = val);
watch(() => props.photoUrl, (val) => newPhotoUrl.value = val || '');

// gestione upload foto
const handleFileUpload = (event) => {
  const file = event.target.files[0];
  if (!file) return;

  // Controllo dimensione (max 1MB circa, dato il limite del backend)
  if (file.size > 1000000) {
    errorMsg.value = "L'immagine è troppo grande (max 1MB).";
    return;
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    newPhotoUrl.value = e.target.result; 
  };
  reader.readAsDataURL(file);
};

// rimozione foto
const removePhoto = () => {
  newPhotoUrl.value = '';
  // Resetta anche l'input file se presente nel DOM
  const fileInput = document.getElementById('fileUploadInput');
  if (fileInput) fileInput.value = '';
};

const handleSave = async () => {
  loading.value = true;
  errorMsg.value = '';
  successMsg.value = '';

  try {
    let updatedName = props.username;
    let updatedPhoto = props.photoUrl;

    // Aggiorna Username
    if (newUsername.value !== props.username) {
      await apiSetMyUserName(newUsername.value);
      updatedName = newUsername.value;
    }

    // Aggiorna Foto
    if (newPhotoUrl.value !== props.photoUrl) {
      await apiSetMyPhoto(newPhotoUrl.value);
      updatedPhoto = newPhotoUrl.value;
    }

    successMsg.value = "Profilo aggiornato con successo!";
    emit('update-profile', { username: updatedName, photoUrl: updatedPhoto });
    
    setTimeout(() => {
      emit('close');
      successMsg.value = '';
    }, 1000);

  } catch (err) {
    errorMsg.value = err.message;
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="modal-title mb-0">Il mio Profilo</h5>
        <button type="button" class="btn-close" @click="emit('close')" />
      </div>
      
      <div class="card-body">
        <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
        <div v-if="successMsg" class="alert alert-success">{{ successMsg }}</div>

        <form @submit.prevent="handleSave">
          
          <div class="mb-3">
            <label class="form-label fw-bold">Username</label>
            <div class="input-group">
              <span class="input-group-text">@</span>
              <input type="text" class="form-control" v-model="newUsername" required minlength="3" maxlength="16">
            </div>
          </div>

          <div class="mb-4">
            <label class="form-label fw-bold">Foto Profilo</label>
            
            <div class="d-flex align-items-center gap-3 mb-2">
              <img 
                :src="newPhotoUrl || 'https://placehold.co/80x80/e9ecef/000000?text=No+Foto'" 
                class="rounded-circle border" 
                width="80" height="80" 
                style="object-fit: cover;" 
                alt="Anteprima"
              >
              
              <div>
                <label class="btn btn-outline-primary btn-sm me-2">
                  Carica Foto...
                  <input 
                    id="fileUploadInput"
                    type="file" 
                    accept="image/*" 
                    class="d-none" 
                    @change="handleFileUpload"
                  >
                </label>
                <button 
                  v-if="newPhotoUrl" 
                  type="button" 
                  class="btn btn-outline-danger btn-sm" 
                  @click="removePhoto"
                >
                  Rimuovi
                </button>
              </div>
            </div>
            
            <div class="form-text">Scegli un'immagine dal tuo dispositivo (max 1MB).</div>
          </div>

          <button type="submit" class="btn btn-primary w-100" :disabled="loading">
            <LoadingSpinner v-if="loading" />
            <span v-else>Salva Modifiche</span>
          </button>
        </form>
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
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
</style>