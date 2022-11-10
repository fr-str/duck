import React from 'react'
import moment from 'moment';
import { useSelector } from 'react-redux'
import CssBaseline from '@mui/material/CssBaseline';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';

const Logs = (props) => {
    // update Components state when props change
    const logs = useSelector(state => state.logs.value)

    let longest = 0;
    // find the longest container name, if log a or b is undefined skip it
    if (logs.length > 0) {
        longest = logs.reduce(function (a, b) {
            return a && a.Container && a.Container.length > b.Container.length ? a : b;
        });
        longest = longest.Container.length
    }
    return (
        <React.Fragment>
            <CssBaseline />
            <Container style={styleC(props.style.height)} sx={{ height: props.style.height}}>
                <Box  />

                {Array.from(logs).map((log, i) => {
                    let pad = "";
                    for (var j = 0; j < longest - log.Container.length; j++) {
                        pad += "\u00A0"
                        // pad += "-"
                    }
                    if (pad.length % 2 !== 0) {
                        pad += "\u00A0"
                    }

                    let paddedName = pad.substring(0, pad.length / 2) + log.Container + pad.substring(0, pad.length / 2)
                    if (paddedName.length > longest) {
                        paddedName = paddedName.substring(0, longest)
                    }

                    let dateAndName = moment(log.Timestamp / 1000 / 1000).format('MMM DD HH:mm:ss:SSS') + " | " + paddedName + " | "
                    return (
                        <div key={i}>
                            <span style={{ color: strToColor(log.Container) }}>{dateAndName}</span>
                            <span style={{ color: strToColor(log.Container) }}>{log.Message}</span>
                        </div>
                    )
                })}
            </Container>
        </React.Fragment>
    )
}

var strToColor = (string, saturation = 100, lightness = 72) => {
    let hash = 0;
    string = string + "?" + string.length
    for (let i = 0; i < string.length; i++) {
        hash = string.charCodeAt(i) + ((hash << 5) - hash);
        hash = hash & hash;
    }
    return `hsl(${(hash % 360)}, ${saturation}%, ${lightness}%)`;
}


// style scrollbars
function styleC(height) {
    return {
        fontFamily: "monospace",
        fontSize: "16px",
        maxWidth: "100vw",
        height: height,
        overflowY: "scroll",
        overflowX: "scroll",
    
        backgroundColor: "#000000",
        overflow: "auto",
        padding: "10px",
        display: "flex",
        flexDirection: "column-reverse",
        borderRadius: "5px",
        whiteSpace: "nowrap",
        color: "#fff",
        position:"fixed",
        bottom:"0px",
    }
}


export default Logs


