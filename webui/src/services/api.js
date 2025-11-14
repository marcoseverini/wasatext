// Leggiamo l'URL del backend dal file .env
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

/*
 * Una funzione "wrapper" per 'fetch' che gestisce la logica del token
 * e imposta gli header corretti per noi.
 */
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

// --- Definizioni delle nostre funzioni API ---

/**
 * Esegue il login e salva il token
 */
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

/**
 * Ottiene la lista delle conversazioni
 */
export async function apiGetMyConversations() {
    return apiFetch('/conversations'); 
}

/**
 * Funzione di Logout
 */
export function apiLogout() {
    localStorage.removeItem('sessionToken');
    window.location.reload();
}

/**
 * Cerca utenti in base al nome
 */
export async function apiSearchUsers(username) {
  return apiFetch(`/users?username=${encodeURIComponent(username)}`);
}

/**
 * Inizia una nuova conversazione 1-a-1
 */
export async function apiStartConversation(userId) {
  return apiFetch('/conversations', {
    method: 'POST',
    body: JSON.stringify({ userId: userId }),
  });
}

/**
 * Ottiene i dettagli completi di una singola conversazione
 */
export async function apiGetConversation(conversationId) {
  return apiFetch(`/conversations/${conversationId}`);
}

/**
 * Invia un nuovo messaggio di testo a una conversazione
 */
export async function apiSendMessage(conversationId, messageText) {
  return apiFetch(`/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({
      text: messageText
    }),
  });
}

/**
 * Crea un nuovo gruppo
 */
export async function apiCreateGroup(groupName, memberIds) {
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
 */
export async function apiDeleteMessage(messageId) {
  return apiFetch(`/messages/${messageId}`, {
    method: 'DELETE',
  });
}