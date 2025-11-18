import { createApp } from 'vue' // Importa la funzione per creare l'app Vue
import App from './App.vue' // Importa il componente principale dell'app
import router from './router' // Importa il router configurato
import ErrorMsg from './components/ErrorMsg.vue' // Importa il componente per i messaggi di errore
import LoadingSpinner from './components/LoadingSpinner.vue' // Importa il componente per lo spinner di caricamento

import 'bootstrap/dist/css/bootstrap.min.css' // Importa lo stile di Bootstrap
import './assets/dashboard.css' // Importa gli stili personalizzati
import './assets/main.css' // Importa altri stili personalizzati

const app = createApp(App) // Crea l'app 

// Componenti globali
app.component("ErrorMsg", ErrorMsg);
app.component("LoadingSpinner", LoadingSpinner);

app.use(router) // Attiva il router
app.mount('#app') // Monta l'app Vue nell'elemento con id 'app'