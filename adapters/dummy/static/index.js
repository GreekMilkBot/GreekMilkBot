let self = fetchObj('/api/self')

function selfInfo() {
    return Object.assign({}, self)
}


function usersInfo() {
    return Object.assign({}, users)
}

function getSession(id) {
    return getSessions()[id];
}

function getSessions() {
    return fetchObj('/api/sessions');
}

function getUserInfo(id) {
   return fetchObj('/api/user?id=' + id)
}

function getGroupInfo(id) {
    return fetchObj('/api/group?id='+ id)
}


function getMessages(id) {
    return  fetchObj('/api/messages?id=' + id)
}


function getMessage(id) {
    return fetchObj('/api/message?id=' + id);
}

function searchUser(session, matchUser) {
    let match = matchUser.toLowerCase()
    let result = []
    if (session.type === 'group') {
        let group = getGroupInfo(session.target);
        for (let [uid, user] of Object.entries(group.users)) {
            if (
                (user.alias && user.alias.toLowerCase().includes(match)) ||
                user.name.toLowerCase().includes(match) ||
                user.id.includes(matchUser)
            ) {
                let name = user.name;
                if (user.alias){
                    name = user.alias
                }
                result.push({
                    id: user.id,
                    avatar: user.avatar,
                    name:name
                })
            }
        }
    }
    return result
}

