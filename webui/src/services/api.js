import axios from "./axios";

// Interceptor: Aggiunge il token a ogni richiesta
axios.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('sessionToken');
        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Interceptor: Gestisce errori globali
axios.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response && error.response.status === 401) {
            localStorage.removeItem('sessionToken');
            localStorage.removeItem('username');
            localStorage.removeItem('photoUrl');
            window.location.href = "/";
        }
        const msg = error.response?.data?.message || error.message;
        return Promise.reject(new Error(msg));
    }
);

// FUNZIONI API

export async function apiLogin(username) {
    const response = await axios.post('/session', { username });
    if (response.data.identifier) {
        localStorage.setItem('sessionToken', response.data.identifier);
        localStorage.setItem('username', username);
        localStorage.removeItem('photoUrl');
    }
    return response.data;
}

export function apiLogout() {
    localStorage.removeItem('sessionToken');
    localStorage.removeItem('username');
    localStorage.removeItem('photoUrl');
    window.location.href = "/";
}

export async function apiGetMyConversations() {
    const response = await axios.get('/conversations');
    return response.data;
}

export async function apiSearchUsers(username) {
    const response = await axios.get('/users', { params: { username } });
    return response.data;
}

export async function apiStartConversation(userId) {
    const response = await axios.post('/conversations', { userId });
    return response.data;
}

export async function apiGetConversation(conversationId) {
    const response = await axios.get(`/conversations/${conversationId}`);
    return response.data;
}

export async function apiSendMessage(conversationId, text, photoUrl, replyToMsgId = null) {
    const payload = {};
    if (replyToMsgId) payload.replyToMsgId = replyToMsgId;
    if (text) payload.text = text;
    if (photoUrl) payload.photoUrl = photoUrl;

    const response = await axios.post(`/conversations/${conversationId}/messages`, payload);
    return response.data;
}

export async function apiCreateGroup(groupName, memberIds) {
    const response = await axios.post('/groups', { groupName, memberIds });
    return response.data;
}

export async function apiDeleteMessage(messageId) {
    const response = await axios.delete(`/messages/${messageId}`);
    return response.data;
}

export async function apiAddReaction(msgId, emoji) {
    const response = await axios.post(`/messages/${msgId}/reactions`, { emoji });
    return response.data;
}

export async function apiRemoveReaction(msgId, reactionId) {
    const response = await axios.delete(`/messages/${msgId}/reactions/${reactionId}`);
    return response.data;
}

export async function apiSetGroupName(convId, name) {
    const response = await axios.put(`/conversations/${convId}/name`, { name });
    return response.data;
}

export async function apiAddToGroup(convId, userId) {
    const response = await axios.post(`/conversations/${convId}/members`, { userId });
    return response.data;
}

export async function apiLeaveGroup(convId) {
    const response = await axios.delete(`/conversations/${convId}/members/me`);
    return response.data;
}

export async function apiSetMyUserName(username) {
    const response = await axios.put('/settings/username', { username });
    return response.data;
}

export async function apiSetMyPhoto(photoUrl) {
    const response = await axios.put('/settings/photo', { photoUrl });
    return response.data;
}

export async function apiSetGroupPhoto(convId, photoUrl) {
    const response = await axios.put(`/conversations/${convId}/photo`, { photoUrl });
    return response.data;
}

export async function apiForwardMessage(targetConvId, originalMsgId) {
    const response = await axios.post(`/conversations/${targetConvId}/forwarded_messages`, { originalMessageId: originalMsgId });
    return response.data;
}