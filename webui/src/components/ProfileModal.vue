<script setup>
import { ref, watch } from 'vue'; // Aggiunto 'watch'
import { apiSetMyUserName, apiSetMyPhoto } from '@/services/api.js';
import ErrorMsg from '@/components/ErrorMsg.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';

const props = defineProps({
  show: Boolean,
  username: String,
  photoUrl: String // <--- NUOVA PROP
});

const emit = defineEmits(['close', 'update-profile']);

const newUsername = ref(props.username);
const newPhotoUrl = ref(props.photoUrl || ''); // Inizializza con la foto attuale
const loading = ref(false);
const errorMsg = ref('');
const successMsg = ref('');

// Aggiorna i campi se le props cambiano mentre il modale è aperto
watch(() => props.username, (val) => newUsername.value = val);
watch(() => props.photoUrl, (val) => newPhotoUrl.value = val || '');

const handleSave = async () => {
  loading.value = true;
  errorMsg.value = '';
  successMsg.value = '';

  try {
    let updatedName = props.username;
    let updatedPhoto = props.photoUrl;

    // 1. Aggiorna Username se cambiato
    if (newUsername.value !== props.username) {
      await apiSetMyUserName(newUsername.value);
      updatedName = newUsername.value;
    }

    // 2. Aggiorna Foto se cambiata
    if (newPhotoUrl.value !== props.photoUrl) {
      await apiSetMyPhoto(newPhotoUrl.value);
      updatedPhoto = newPhotoUrl.value;
    }

    successMsg.value = "Profilo aggiornato con successo!";
    
    // Emette entrambi i nuovi valori al padre (App.vue)
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
            <label class="form-label fw-bold">URL Foto Profilo</label>
            <input type="url" class="form-control" v-model="newPhotoUrl" placeholder="https://...">
            <div class="form-text">Incolla un link a un'immagine. Lascia vuoto per rimuoverla.</div>
          </div>

          <div v-if="newPhotoUrl" class="text-center mb-3">
            <label class="form-label small text-muted d-block">Anteprima</label>
            <img :src="newPhotoUrl" class="rounded-circle border" width="80" height="80" style="object-fit: cover;" alt="Anteprima">
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