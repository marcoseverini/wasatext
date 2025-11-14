// Leggiamo l'URL del backend dal file .env
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// Una funzione che gestisce la logica del token e imposta gli header corretti per noi.
async function apiFetch(endpoint, options = {}) {

    // Prepara gli header
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    // Leggi il token salvato dal localStorage
    const token = localStorage.getItem('sessionToken');
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    // Costruisci la richiesta
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: headers,
    });

    // Se la risposta è 401 (token non valido/scaduto),
    // cancella il token e ricarica la pagina (che forzerà il login)
    if (response.status === 401) {
        localStorage.removeItem('sessionToken');
        window.location.reload();
        throw new Error("Sessione scaduta. Effettua nuovamente il login.");
    }

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Si è verificato un errore');
    }

    if (response.status === 204) {
        return null;
    }

    return response.json();
}

// Definizioni delle nostre funzioni API

// Esegue il login e salva il token
export async function apiLogin(username) {
    const response = await apiFetch('/session', {
        method: 'POST',
        body: JSON.stringify({ username: username }), 
    });

    if (response.identifier) {
        localStorage.setItem('sessionToken', response.identifier);
    }
    return response;
}

// Ottiene la lista delle conversazioni
export async function apiGetMyConversations() {
    return apiFetch('/conversations'); 
}

// Rimuove il token e ricarica la pagina.
export function apiLogout() {
    localStorage.removeItem('sessionToken');
    // Ricaricando, il router guard (beforeEach) ci riporterà
    // automaticamente alla pagina di login.
    window.location.reload();
}

// (Aggiungeremo le altre qui quando serviranno)