import Containers from './components/Containers';
import Logs from './components/Logs';
import { Connect ,Live } from './Websocket/Websocket';
import React from 'react'
import { useState } from 'react'
import styled from 'styled-components'
import { setConts } from './app/ws'
import store from './app/store'

Connect()

const App = () => {
  // read localstorage includeContainers if empty set to empty array
  const includeContainers = JSON.parse(localStorage.getItem('includeContainers')) || []
  // set includeContainers state
  // if includeContainers is not empty, send request to websocket
  if (includeContainers.length > 0) {
    store.dispatch(setConts(includeContainers))
    Live()
  }

  const [size, setSize] = useState({ y: 350 });

  const handler = (mouseDownEvent) => {
    const startSize = size;
    const startPosition = { y: mouseDownEvent.pageY };
    
    function onMouseMove(mouseMoveEvent) {
      // show ghost 
      mouseMoveEvent.preventDefault();
    }
    function onMouseUp(mousePosition) {
      document.body.removeEventListener("mousemove", onMouseMove);
      setSize(currentSize => ({ 
        y: startSize.y - startPosition.y + mousePosition.pageY 
      }));
    }
    
    document.body.addEventListener("mousemove", onMouseMove);
    document.body.addEventListener("mouseup", onMouseUp, { once: true });
  };

  return (
    <AppDiv >
      <Containers style={{ height: size.y }} />
      <button style={{ position: 'absolute', bottom: window.innerHeight-size.y, right: 0, width: '100%', height: '10px', cursor: 'ns-resize', backgroundColor: 'grey' }} onMouseDown={handler} />
      <Logs style={{ height: window.innerHeight-size.y }} />
    </AppDiv>
  );
}

export default App;

// change background color

document.body.style.backgroundColor = "#1f2223";

// App styled div
const AppDiv = styled.div`
    height: 100%;
    width: 100%;
    background-color: #1f2223;
    font-size: 20px;
    display: grid;
    min-height: 100vh;
    overflow: hidden;    
`