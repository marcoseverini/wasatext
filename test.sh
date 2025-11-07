#!/bin/bash

# Script di test di integrazione per l'API WASAText
# Eseguire con: bash test-all.sh

# --- Configurazione ---
BASE_URL="http://localhost:3000"
JQ_CMD="jq" # Richiede 'jq' installato (es. 'apt install jq')

# --- Funzioni Helper per i colori ---
COLOR_GREEN="\033[0;32m"
COLOR_RED="\033[0;31m"
COLOR_YELLOW="\033[0;33m"
COLOR_NONE="\033[0m"

echo_ok()    { echo -e "${COLOR_GREEN}✅ $1${COLOR_NONE}"; }
echo_fail()  { echo -e "${COLOR_RED}❌ $1${COLOR_NONE}"; }
echo_info()  { echo -e "${COLOR_YELLOW}ℹ️ $1${COLOR_NONE}"; }

# Funzione per controllare lo status code
assert_status() {
    local response=$1
    local expected_status=$2
    local test_name=$3
    
    local status=$(echo "$response" | grep "HTTP/1.1" | awk '{print $2}')
    
    if [ "$status" == "$expected_status" ]; then
        echo_ok "$test_name (Status $status)"
    else
        echo_fail "$test_name (Atteso $expected_status, ricevuto $status)"
        echo "$response" # Stampa l'errore
        exit 1 # Interrompe il test
    fi
}

# Funzione per estrarre il JSON dal corpo della risposta curl -v
extract_body() {
    echo "$1" | sed -n '/^{/,$p'
}

# Assicurati che il server sia raggiungibile
echo_info "Ping del server su $BASE_URL..."
curl -s --head $BASE_URL/session > /dev/null
if [ $? -ne 0 ]; then
    echo_fail "Il server non è in esecuzione su $BASE_URL. Avvialo prima di eseguire i test."
    exit 1
fi

echo_info "--- Inizio Test di Integrazione ---"

# ===============================================
echo_info "Blocco 1: Login e Setup Utenti"
# ===============================================

# 1.1 Test Login (Maria) - 201
echo_info "Test 1.1: Login 'Maria' (201)"
RES_MARIA_RAW=$(curl -s -v -X POST "$BASE_URL/session" -H "Content-Type: application/json" -d '{"username": "Maria"}')
assert_status "$RES_MARIA_RAW" "201" "Login Maria"
JSON_MARIA=$(extract_body "$RES_MARIA_RAW")
TOKEN_MARIA=$(echo $JSON_MARIA | $JQ_CMD -r .identifier)
ID_MARIA=$(echo $JSON_MARIA | $JQ_CMD -r .identifier)

# 1.2 Test Login (Luca) - 201
echo_info "Test 1.2: Login 'Luca' (201)"
RES_LUCA_RAW=$(curl -s -v -X POST "$BASE_URL/session" -H "Content-Type: application/json" -d '{"username": "Luca"}')
assert_status "$RES_LUCA_RAW" "201" "Login Luca"
JSON_LUCA=$(extract_body "$RES_LUCA_RAW")
TOKEN_LUCA=$(echo $JSON_LUCA | $JQ_CMD -r .identifier)
ID_LUCA=$(echo $JSON_LUCA | $JQ_CMD -r .identifier)

# 1.3 Test Login (Carlo) - 201
echo_info "Test 1.3: Login 'Carlo' (201)"
RES_CARLO_RAW=$(curl -s -v -X POST "$BASE_URL/session" -H "Content-Type: application/json" -d '{"username": "Carlo"}')
assert_status "$RES_CARLO_RAW" "201" "Login Carlo"
TOKEN_CARLO=$(echo $JSON_CARLO | $JQ_CMD -r .identifier)

# 1.4 Test Login (Nome breve) - 400
echo_info "Test 1.4: Login 'io' (400 Bad Request)"
RES_ERR_RAW=$(curl -s -v -X POST "$BASE_URL/session" -H "Content-Type: application/json" -d '{"username": "io"}')
assert_status "$RES_ERR_RAW" "400" "Login nome breve"

# ===============================================
echo_info "Blocco 2: Settings (setMyUsername, setMyPhoto)"
# ===============================================

# 2.1 Test setMyUsername (Successo) - 200
echo_info "Test 2.1: setMyUsername (200 OK)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/settings/username" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"username": "Maria99"}')
assert_status "$RES_RAW" "200" "setMyUsername"

# 2.2 Test setMyUsername (Conflitto) - 409
echo_info "Test 2.2: setMyUsername - Conflitto (409 Conflict)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/settings/username" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"username": "Luca"}') # Prova a prendere il nome di Luca
assert_status "$RES_RAW" "409" "setMyUsername Conflitto"

# 2.3 Test setMyPhoto (Successo) - 200
echo_info "Test 2.3: setMyPhoto (200 OK)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/settings/photo" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"photoUrl": "https://example.com/foto.png"}')
assert_status "$RES_RAW" "200" "setMyPhoto"

# 2.4 Test setMyPhoto (URL non valido) - 400
echo_info "Test 2.4: setMyPhoto - URL non valido (400 Bad Request)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/settings/photo" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"photoUrl": "non-un-url"}')
assert_status "$RES_RAW" "400" "setMyPhoto URL non valido"

# 2.5 Test setMyUsername (Senza token) - 401
echo_info "Test 2.5: setMyUsername - Senza token (401 Unauthorized)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/settings/username" \
    -H "Content-Type: application/json" -d '{"username": "MariaFAIL"}')
assert_status "$RES_RAW" "401" "setMyUsername senza token"

# ===============================================
echo_info "Blocco 3: Users (searchUsers)"
# ===============================================

# 3.1 Test searchUsers (Successo) - 200
echo_info "Test 3.1: searchUsers 'Luca' (200 OK)"
RES_RAW=$(curl -s -v -X GET "$BASE_URL/users?username=Luca" \
    -H "Authorization: Bearer $TOKEN_MARIA")
assert_status "$RES_RAW" "200" "searchUsers"
BODY=$(extract_body "$RES_RAW")
echo "Risultati trovati: $BODY"

# 3.2 Test searchUsers (Senza token) - 401
echo_info "Test 3.2: searchUsers (401 Unauthorized)"
RES_RAW=$(curl -s -v -X GET "$BASE_URL/users?username=Luca")
assert_status "$RES_RAW" "401" "searchUsers senza token"

# 3.3 Test searchUsers (Query non valida) - 400
echo_info "Test 3.3: searchUsers query vuota (400 Bad Request)"
RES_RAW=$(curl -s -v -X GET "$BASE_URL/users?username=" \
    -H "Authorization: Bearer $TOKEN_MARIA")
assert_status "$RES_RAW" "400" "searchUsers query non valida"

# ===============================================
echo_info "Blocco 4: Conversazioni (1-a-1 e Messaggi)"
# ===============================================

# 4.1 Test startConversation (Successo) - 201
echo_info "Test 4.1: startConversation Maria+Luca (201 Created)"
RES_CONV_RAW=$(curl -s -v -X POST "$BASE_URL/conversations" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d "{\"userId\": \"$ID_LUCA\"}")
assert_status "$RES_CONV_RAW" "201" "startConversation"
CONV_ID_1=$(extract_body "$RES_CONV_RAW" | $JQ_CMD -r .id)
echo "ID Conversazione 1: $CONV_ID_1"

# 4.2 Test startConversation (Utente non trovato) - 404
echo_info "Test 4.2: startConversation (404 Not Found)"
RES_RAW=$(curl -s -v -X POST "$BASE_URL/conversations" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"userId": "123e4567-e89b-12d3-a456-426614174000"}')
assert_status "$RES_RAW" "404" "startConversation utente 404"

# 4.3 Test sendMessage (Successo) - 201
echo_info "Test 4.3: sendMessage (201 Created)"
RES_MSG_RAW=$(curl -s -v -X POST "$BASE_URL/conversations/$CONV_ID_1/messages" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"text": "Ciao Luca!"}')
assert_status "$RES_MSG_RAW" "201" "sendMessage"
MSG_ID_1=$(extract_body "$RES_MSG_RAW" | $JQ_CMD -r .id)
echo "ID Messaggio 1: $MSG_ID_1"

# 4.4 Test sendMessage (Errore 403) - 403
echo_info "Test 4.4: sendMessage (403 Forbidden)"
RES_RAW=$(curl -s -v -X POST "$BASE_URL/conversations/$CONV_ID_1/messages" \
    -H "Authorization: Bearer $TOKEN_CARLO" \
    -H "Content-Type: application/json" -d '{"text": "Intruso!"}')
assert_status "$RES_RAW" "403" "sendMessage utente 403"

# 4.5 Test getConversation (Successo) - 200
echo_info "Test 4.5: getConversation (200 OK)"
RES_RAW=$(curl -s -v -X GET "$BASE_URL/conversations/$CONV_ID_1" \
    -H "Authorization: Bearer $TOKEN_LUCA") # Luca controlla la chat
assert_status "$RES_RAW" "200" "getConversation"

# 4.6 Test getMyConversations (Successo) - 200
echo_info "Test 4.6: getMyConversations (200 OK)"
RES_RAW=$(curl -s -v -X GET "$BASE_URL/conversations" \
    -H "Authorization: Bearer $TOKEN_MARIA")
assert_status "$RES_RAW" "200" "getMyConversations"
BODY=$(extract_body "$RES_RAW")
echo "Lista chat di Maria: $BODY"

# ===============================================
echo_info "Blocco 5: Reazioni e Delete Messaggi"
# ===============================================

# 5.1 Test commentMessage (Successo) - 201
echo_info "Test 5.1: commentMessage (201 Created)"
RES_REACT_RAW=$(curl -s -v -X POST "$BASE_URL/messages/$MSG_ID_1/reactions" \
    -H "Authorization: Bearer $TOKEN_LUCA" \
    -H "Content-Type: application/json" -d '{"emoji": "👍"}')
assert_status "$RES_REACT_RAW" "201" "commentMessage"
REACT_ID=$(extract_body "$RES_REACT_RAW" | $JQ_CMD -r .id)
echo "ID Reazione: $REACT_ID"

# 5.2 Test uncommentMessage (Errore 403) - 403
echo_info "Test 5.2: uncommentMessage (403 Forbidden)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/messages/$MSG_ID_1/reactions/$REACT_ID" \
    -H "Authorization: Bearer $TOKEN_MARIA") # Maria prova a togliere la reazione di Luca
assert_status "$RES_RAW" "403" "uncommentMessage utente 403"

# 5.3 Test uncommentMessage (Successo) - 204
echo_info "Test 5.3: uncommentMessage (204 No Content)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/messages/$MSG_ID_1/reactions/$REACT_ID" \
    -H "Authorization: Bearer $TOKEN_LUCA") # Luca toglie la sua
assert_status "$RES_RAW" "204" "uncommentMessage"

# 5.4 Test deleteMessage (Errore 403) - 403
echo_info "Test 5.4: deleteMessage (403 Forbidden)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/messages/$MSG_ID_1" \
    -H "Authorization: Bearer $TOKEN_LUCA") # Luca prova a cancellare il messaggio di Maria
assert_status "$RES_RAW" "403" "deleteMessage utente 403"

# 5.5 Test deleteMessage (Successo) - 204
echo_info "Test 5.5: deleteMessage (204 No Content)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/messages/$MSG_ID_1" \
    -H "Authorization: Bearer $TOKEN_MARIA") # Maria cancella il suo
assert_status "$RES_RAW" "204" "deleteMessage"

# ===============================================
echo_info "Blocco 6: Gruppi"
# ===============================================

# 6.1 Test createGroup (Successo) - 201
echo_info "Test 6.1: createGroup Maria+Luca (201 Created)"
RES_GROUP_RAW=$(curl -s -v -X POST "$BASE_URL/groups" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d "{\"groupName\": \"Gruppo Test\", \"memberIds\": [\"$ID_LUCA\"]}")
assert_status "$RES_GROUP_RAW" "201" "createGroup"
GROUP_ID=$(extract_body "$RES_GROUP_RAW" | $JQ_CMD -r .id)
echo "ID Gruppo: $GROUP_ID"

# 6.2 Test setGroupName (Successo) - 200
echo_info "Test 6.2: setGroupName (200 OK)"
RES_RAW=$(curl -s -v -X PUT "$BASE_URL/conversations/$GROUP_ID/name" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d '{"name": "Nuovo Nome Gruppo"}')
assert_status "$RES_RAW" "200" "setGroupName"

# 6.3 Test addToGroup (Successo) - 204
echo_info "Test 6.3: addToGroup - Aggiungi Carlo (204 No Content)"
RES_RAW=$(curl -s -v -X POST "$BASE_URL/conversations/$GROUP_ID/members" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d "{\"userId\": \"$ID_CARLO\"}")
assert_status "$RES_RAW" "204" "addToGroup"

# 6.4 Test addToGroup (Conflitto) - 409
echo_info "Test 6.4: addToGroup - Aggiungi Carlo di nuovo (409 Conflict)"
RES_RAW=$(curl -s -v -X POST "$BASE_URL/conversations/$GROUP_ID/members" \
    -H "Authorization: Bearer $TOKEN_MARIA" \
    -H "Content-Type: application/json" -d "{\"userId\": \"$ID_CARLO\"}")
assert_status "$RES_RAW" "409" "addToGroup Conflitto"

# 6.5 Test leaveGroup (Successo) - 204
echo_info "Test 6.5: leaveGroup - Carlo esce (204 No Content)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/conversations/$GROUP_ID/members/me" \
    -H "Authorization: Bearer $TOKEN_CARLO")
assert_status "$RES_RAW" "204" "leaveGroup"

# 6.6 Test leaveGroup (Errore 403) - 403
echo_info "Test 6.6: leaveGroup - Carlo prova a uscire di nuovo (403 Forbidden)"
RES_RAW=$(curl -s -v -X DELETE "$BASE_URL/conversations/$GROUP_ID/members/me" \
    -H "Authorization: Bearer $TOKEN_CARLO")
assert_status "$RES_RAW" "403" "leaveGroup utente 403"

echo_ok "--- Tutti i test sono stati superati! ---"