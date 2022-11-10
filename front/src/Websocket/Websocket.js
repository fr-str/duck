
//import { useDispatch } from 'react-redux'
import { remContainer, setContainer } from '../app/containers'
import { popLog, addLogs,clearLogs } from '../app/logs'
import { remCont } from '../app/ws'
import store from '../app/store'
export var WS = null;

export function Connect() {

    WS = new WebSocket("wss://docker-project.dodupy.dev/api");
    WS.onopen = () => {
        console.log("Connected to websocket");
        Live();

    };
    WS.onmessage = (msg) => {
        const data = JSON.parse(msg.data)
        // update containers map state
        // console.log(data)
        if (data.Code === 200 && data.RequestID === 'containers') {
            const { Event, Key, Value } = data.Data
            // console.log(Event,Key,Value)
            if (Event === 'PUT') {
                store.dispatch(setContainer({ key: Key, value: Value }))
            }

            if (Event === 'DELETE') {
                store.dispatch(remContainer(Key))
            }
        }

        // append logs to logs state, max 5000 logs
        if (data.Code === 200 && data.RequestID === 'logs') {
            // console.log(data.Data)
            store.dispatch(addLogs(data.Data))
            if (data.Data.length > 5000) {
                store.dispatch(popLog())
            }
        }
        if (data.Code === 200 && data.RequestID === 'inspect') {
            console.log(data.Data)
        }
        if (data.Code === 311 && data.RequestID === 'live') {
            store.dispatch(remCont(data.Data))
        }
    }

    WS.onclose = () => {
        console.log("Disconnected from websocket");
        setTimeout(Connect, 500);
    };
}

export async function Live() {
    if (WS.readyState !== 1) {
        setTimeout(Live, 100);
        return;
    }

    WS.send(JSON.stringify({
        "RequestID": "live",
        "Action": "live",
        "Data": {
            "Containers": {},
            "Logs": {
                "ContainerNames": store.getState("includeContainers").includeContainers.value,
                "Amount": 100
            }
        }
    }));
    store.dispatch(clearLogs(store.getState("includeContainers").includeContainers.value))
}

export async function Send(rID, action, data) {
    if (WS.readyState !== 1) {
        setTimeout(() => Send(rID, action, data), 1000);
        return;
    }


    WS.send(JSON.stringify({
        "RequestID": rID,
        "Action": action,
        "Data": data
      }));
}