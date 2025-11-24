import axios from "axios";

// Creazione dell'istanza Axios
const api = axios.create({
	baseURL: __API_URL__, // Questa variabile viene iniettata da Vite/Go in fase di build
	timeout: 1000 * 5, // 5 secondi di timeout
});

// Interceptor per aggiungere l'Authorization header a ogni richiesta
// (prende il token salvato nel localStorage)
api.interceptors.request.use(
	(config) => {
		const token = localStorage.getItem("sessionToken");
		if (token) {
			config.headers["Authorization"] = `Bearer ${token}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

// --- FUNZIONI API ---

// Login
export const apiLogin = async (username) => {
	// POST /session
	const response = await api.post("/session", { username });
	// Salva l'ID utente (che funge da token)
	localStorage.setItem("sessionToken", response.data.identifier);
	localStorage.setItem("username", username);
	return response.data;
};

// Logout (semplicemente rimuove il token locale)
export const apiLogout = () => {
	localStorage.removeItem("sessionToken");
	localStorage.removeItem("username");
	window.location.href = "/"; // Ricarica la pagina per tornare al login
};

// Ottieni le mie conversazioni
export const apiGetMyConversations = async () => {
	// GET /conversations
	const response = await api.get("/conversations");
	return response.data;
};

// Ottieni una singola conversazione e i suoi messaggi
export const apiGetConversation = async (convId) => {
	// GET /conversations/{convId}
	const response = await api.get(`/conversations/${convId}`);
	return response.data;
};

// Cerca utenti
export const apiSearchUsers = async (query) => {
	// GET /users?username=...
	const response = await api.get("/users", {
		params: { username: query },
	});
	return response.data;
};

// Inizia una nuova chat
export const apiStartConversation = async (recipientId) => {
	// POST /conversations
	const response = await api.post("/conversations", {
		userId: recipientId, // NOTA: Nello YAML è 'userId' dentro UserIdRequest
	});
	return response.data;
};

// Invia Messaggio
export const apiSendMessage = async (convId, text) => {
	// POST /conversations/{convId}/messages
	const response = await api.post(`/conversations/${convId}/messages`, {
		text: text,
	});
	return response.data;
};

// Cancella Messaggio
export const apiDeleteMessage = async (msgId) => {
	// DELETE /messages/{msgId}
	await api.delete(`/messages/${msgId}`);
};

// Aggiungi Reazione
export const apiAddReaction = async (msgId, emoji) => {
	// POST /messages/{msgId}/reactions
	const response = await api.post(`/messages/${msgId}/reactions`, { emoji });
	return response.data;
};

// Rimuovi Reazione
export const apiRemoveReaction = async (msgId, reactionId) => {
	// DELETE /messages/{msgId}/reactions/{reactionId}
	await api.delete(`/messages/${msgId}/reactions/${reactionId}`);
};