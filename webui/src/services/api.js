// Leggiamo l'URL del backend dal file .env (Vite lo inietta qui)
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

    // Se c'è un errore (status non 2xx), lancia un'eccezione
    if (!response.ok) {
        let errorData;
        try {
            errorData = await response.json();
        } catch (e) {
            errorData = {};
        }
        throw new Error(errorData.message || 'Si è verificato un errore: ' + response.statusText);
    }

    // Se è 204 No Content, ritorna null
    if (response.status === 204) {
        return null;
    }
    
    // Altrimenti ritorna il JSON
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
        // Salviamo anche l'username per comodità nel frontend
        localStorage.setItem('username', username);
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
    localStorage.removeItem('username');
    window.location.href = "/";
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
 * Accetta un replyToMsgId opzionale
 */
export async function apiSendMessage(conversationId, messageText, replyToMsgId = null) {
  const payload = { text: messageText };
  if (replyToMsgId) {
    payload.replyToMsgId = replyToMsgId;
  }

  return apiFetch(`/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify(payload),
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

/**
 * Aggiunge una reazione a un messaggio (CORRETTA)
 */
export async function apiAddReaction(msgId, emoji) {
    return apiFetch(`/messages/${msgId}/reactions`, {
        method: 'POST',
        body: JSON.stringify({ emoji: emoji })
    });
}

/**
 * Rimuove una reazione (CORRETTA)
 */
export async function apiRemoveReaction(msgId, reactionId) {
    return apiFetch(`/messages/${msgId}/reactions/${reactionId}`, {
        method: 'DELETE'
    });
}

// Aggiorna il nome del gruppo
export async function apiSetGroupName(convId, name) {
    return apiFetch(`/conversations/${convId}/name`, {
        method: 'PUT',
        body: JSON.stringify({ name })
    });
}

// Aggiunge un utente a un gruppo esistente
export async function apiAddToGroup(convId, userId) {
    // POST /conversations/{convId}/members
    return apiFetch(`/conversations/${convId}/members`, {
        method: 'POST',
        body: JSON.stringify({ userId })
    });
}

// Abbandona il gruppo
export async function apiLeaveGroup(convId) {
    // DELETE /conversations/{convId}/members/me
    return apiFetch(`/conversations/${convId}/members/me`, {
        method: 'DELETE'
    });
}

// Imposta il proprio username
export async function apiSetMyUserName(username) {
    // PUT /settings/username
    return apiFetch('/settings/username', {
        method: 'PUT',
        body: JSON.stringify({ username })
    });
}

// Imposta la propria foto profilo
export async function apiSetMyPhoto(photoUrl) {
    // PUT /settings/photo
    return apiFetch('/settings/photo', {
        method: 'PUT',
        body: JSON.stringify({ photoUrl })
    });
}