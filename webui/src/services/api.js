// Legge l'URL del backend dal file .env
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// Gestisce la logica del token e imposta gli header corretti.
async function apiFetch(endpoint, options = {}) {

    // Prepara gli header
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    // Legge il token salvato dal localStorage
    const token = localStorage.getItem('sessionToken');
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    // Costruisce la richiesta
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

// Definizione delle funzioni API

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

/**
 * Cerca utenti in base al nome
 * @param {string} username - Il termine di ricerca
 * @returns {Promise<object>} La lista degli utenti
 */
export async function apiSearchUsers(username) {
  // Costruisce la query string e chiama GET /users
  return apiFetch(`/users?username=${encodeURIComponent(username)}`);
}

/**
 * Inizia una nuova conversazione 1-a-1
 * @param {string} userId - L'ID dell'utente con cui chattare
 * @returns {Promise<object>} La nuova conversazione
 */
export async function apiStartConversation(userId) {
  // Chiama POST /conversations (corrisponde a UserIdRequest)
  return apiFetch('/conversations', {
    method: 'POST',
    body: JSON.stringify({ userId: userId }),
  });
}

/**
 * Ottiene i dettagli completi di una singola conversazione
 * @param {string} conversationId - L'ID della conversazione
 * @returns {Promise<object>} L'oggetto conversazione (con messaggi, membri, ecc.)
 */
export async function apiGetConversation(conversationId) {
  return apiFetch(`/conversations/${conversationId}`);
}

/**
 * Invia un nuovo messaggio di testo a una conversazione
 * @param {string} conversationId - L'ID della conversazione
 * @param {string} messageText - Il testo da inviare
 * @returns {Promise<object>} Il nuovo oggetto messaggio creato
 */
export async function apiSendMessage(conversationId, messageText) {
  // Corrisponde allo schema SendMessageRequest (con 'text')
  return apiFetch(`/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({
      text: messageText
    }),
  });
}

/**
 * Crea un nuovo gruppo
 * @param {string} groupName - Il nome del nuovo gruppo
 * @param {string[]} memberIds - Un array di ID utente da includere
 * @returns {Promise<object>} La nuova conversazione di gruppo
 */
export async function apiCreateGroup(groupName, memberIds) {
  // Corrisponde a CreateGroupRequest
  return apiFetch('/groups', {
    method: 'POST',
    body: JSON.stringify({
      groupName: groupName,
      memberIds: memberIds,
    }),
  });
}

/**
 * Cancella un messaggio
 * @param {string} messageId - L'ID del messaggio da cancellare
 * @returns {Promise<null>} Una promessa che si risolve (con null) se ha successo
 */
export async function apiDeleteMessage(messageId) {
  // Chiama l'endpoint DELETE. 
  // Il nostro 'apiFetch' gestisce già la risposta 204 No Content.
  return apiFetch(`/messages/${messageId}`, {
    method: 'DELETE',
  });
}

// (Aggiungeremo le altre qui quando serviranno)