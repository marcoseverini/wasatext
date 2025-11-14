import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

// Importa gli stili di Bootstrap (dalla cartella public) e gli stili
import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App) 

// Componenti globali
app.component("ErrorMsg", ErrorMsg);
app.component("LoadingSpinner", LoadingSpinner);

app.use(router) // Attiva il router
app.mount('#app')