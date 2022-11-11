import * as React from 'react';
import moment from 'moment';
import { addCont, remCont } from '../app/ws'
import { clearInspect } from '../app/containers'
import { useSelector } from 'react-redux'
import { Live, Send } from '../Websocket/Websocket'
import store from '../app/store'
import { styled } from '@mui/material/styles';
import Table, { tableClasses } from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { tableCellClasses } from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper, { paperClasses } from '@mui/material/Paper';
import Button from '@mui/material/Button';
import { createTheme } from '@mui/material/styles';
import Popover, { popoverClasses } from '@mui/material/Popover';
import Typography, { typographyClasses } from '@mui/material/Typography';
import Modal from '@mui/material/Modal';
import Fade from '@mui/material/Fade';
import Backdrop from '@mui/material/Backdrop';
import Box from '@mui/material/Box';
import {JSONTree} from 'react-json-tree'


const Containers = (props) => {
  // update Components state when props change
  const containers = useSelector(state => state.containers.value)
  // eslint-disable-next-line
  const _ = useSelector(state => state.includeContainers.value)
  // console.log(containers)
  return (
    <TableContainer component={StyledPaper}>
      <StyledTable padding='none' sx={{ minWidth: 700, maxHeight: props.style.height }} aria-label="customized table">
        <TableHead >
          <TableRow >
            <StyledTableCell sx={{p:2}} align="left">Container Name</StyledTableCell>
            <StyledTableCell sx={{p:2}} align="left">Status</StyledTableCell>
            <StyledTableCell sx={{p:2}} align="left">Image</StyledTableCell>
            <StyledTableCell sx={{p:2}} align="left">Crated</StyledTableCell>
            <StyledTableCell sx={{p:2}} align="left">Metrics</StyledTableCell>
            <StyledTableCell sx={{p:2}} align="center">Actions</StyledTableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {Array.from(containers).sort().map((cont) => {
            let onClick = () => { store.dispatch(addCont(cont[1].Name)); Live() }
            let variant = "contained"
            let name = "Tail"

            if (store.getState().includeContainers.value.includes(cont[1].Name)) {
              variant = "outlined"
              onClick = () => { store.dispatch(remCont(cont[1].Name)); Live() }
              name = "Tailing..."
            }

            return (
              <StyledTableRow sx={{ p: 1 }} key={cont[1].Name}>
                <StyledTableCell component="th" scope="cont">
                  {cont[1].Name}
                </StyledTableCell>
                <StyledTableCell component="th" scope="cont">{cont[1].Status}</StyledTableCell>
                <StyledTableCell component="th" scope="cont">{cont[1].Image}</StyledTableCell>
                <StyledTableCell component="th" scope="cont">{moment.unix(cont[1].Created).fromNow()}</StyledTableCell>
                <StyledTableCell component="th" scope="cont">Nie ma</StyledTableCell>
                <StyledTableCell component="th" scope="cont" align="right" style={{ paddingRight: 15 }}>

                  <Button
                    style={buttonStyle}
                    variant={variant}
                    onClick={onClick}
                  >{name}</Button>


                  <Inspect cont={cont[1]} />
                  {/* <Button 
                  style={buttonStyle} 
                  variant="contained" 
                  onClick={() => { getDetails(cont[0]) }}
                  >Inspect</Button> */}

                  <PopoverButton
                    style={buttonStyle}
                    theme={restartTheme}
                    text="Restart"
                    func={() => { Send("restart" + cont[0], "container.restart", { "Name": cont[0] }) }}
                  />

                  <StartStopButton
                    cont={cont[1]}
                  />

                  <PopoverButton
                    style={buttonStyle}
                    theme={killTheme}
                    text="kill"
                    disabled={cont[1].Status.includes("Exited") ? true : false}
                    func={() => {  Send("kill" + cont[0], "container.kill", { "Name": cont[0] }) }} />

                </StyledTableCell>
              </StyledTableRow>
            )
          })}
        </TableBody>
      </StyledTable>
    </TableContainer>
  );
}

function PopoverButton(props) {
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const open = Boolean(anchorEl);
  const id = open ? 'simple-popover' : undefined;
  return (
    <span>
      <Button
        style={props.style}
        theme={props.theme}
        aria-describedby={id}
        variant="contained"
        // disabled={props.disabled}
        onClick={handleClick}>
        {props.text}
      </Button>
      <StyledPopover
        PaperProps={{
          style: {
            backgroundColor: '#FFA836',
            boxShadow: 'none',
            borderRadius: 3,
          },
        }}
        // style={props.style}
        theme={props.theme}
        id={id}
        open={open}
        anchorEl={anchorEl}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
      >
        <StyledPopoverTypography sx={{ p: 2 }}>
          <Button variant="contained"
            style={{ background: props.theme.palette.primary.main }}
            onClick={
              () => {
                props.func()
                handleClose()
              }
            }>Confirm</Button>
        </StyledPopoverTypography>
      </StyledPopover>
    </span>
  );
}

function StartStopButton(props) {
  //watch for changes in props
  const cont = props.cont

  if (cont.Status.includes("Up")) {
    return (
      <PopoverButton
        style={buttonStyle}
        theme={stopTheme}
        text="Stop"
        func={() => { Send("stop" + cont.Name, "container.stop", { "Name": cont.Name }) }}
      />
    )
  }
  return (
    <Button
      style={buttonStyle}
      variant="contained"
      onClick={() => { Send("start" + cont.Name, "container.start", { "Name": cont.Name }) }}
      color={"success"}
      sx={{ width: 79 }}>
      Start
    </Button>
  )

}

// modal
function Inspect(props) {
  const [open, setOpen] = React.useState(false);
  const handleOpen = () => {
    Send("inspect", "container.inspect", { Name: props.cont.Name })
    setOpen(true)
  };
  const handleClose = () => {
    setOpen(false)
    setTimeout(() => {
      store.dispatch(clearInspect())
    }, 100);
  };

  return (
    <span>
      <Button style={buttonStyle} variant="contained" onClick={handleOpen}>Inspect</Button>
      <StyledModal handleClose={handleClose} open={open} />
    </span>
  );
}

function StyledModal(props) {
  const data = useSelector(state => state.inspect.value)
  return (
    <Modal
      aria-labelledby="transition-modal-title"
      aria-describedby="transition-modal-description"
      open={props.open}
      onClose={props.handleClose}
      closeAfterTransition
      BackdropComponent={Backdrop}
      BackdropProps={{
        timeout: 500,
      }}
    >
      <Fade in={props.open}>
        <Box sx={boxStyle}>
          <Typography
            style={{ backgroundColor:'#002b36', color: "white", fontSize: 14, fontWeight: 600, marginBottom: 10 }}
            id="transition-modal-title"
            variant="h6"
            component="h2">
            <JSONTree
              data={Object.assign({}, data, { children: 'function () {...}' })}
              invertTheme
            />
          </Typography>
        </Box>
      </Fade>
    </Modal>
  )
}

const boxStyle = {
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate(-50%, -50%)',
  width: '80%',
  maxHeight: '80%',
  color: 'white',
  bgcolor: '#002b36',
  border: '2px solid #000',
  boxShadow: 24,
  p: 4,
  overflowX: 'auto',
  overflowY: 'auto',

};

//kill theme
const killTheme = createTheme({
  palette: {
    primary: {
      main: '#a91409',
    },
  },
});

//restart theme
const restartTheme = createTheme({
  palette: {
    primary: {
      main: '#ed6c02',
    },
  },
});

//stop theme
const stopTheme = createTheme({
  palette: {
    primary: {
      main: '#d32f2f',
    },
  },
});

//button style height 20px
const buttonStyle = {
  height: 35,
  padding: 0,
  margin: 0,
  minWidth: 0,
  width: 79,
  borderRadius: 0,
}

const StyledTableCell = styled(TableCell)(({ theme }) => ({
  [`&.${tableCellClasses.head}`]: {
    backgroundColor: theme.palette.common.black,
    color: theme.palette.common.white,
  },

  [`&.${tableCellClasses.body}`]: {
    color: theme.palette.common.white,
    fontSize: 16,
  },
}));

const StyledTableRow = styled(TableRow)(({ theme }) => ({
  '&:nth-of-type(odd)': {
    backgroundColor: theme.palette.action.hover,
  },
  '& td, & th': {
    borderBottom: '1px solid #393d40',
    paddingLeft: 20,
  },
  '&:last-child td, &:last-child th': {
    border: 0,
  },
}));


//styled table
const StyledTable = styled(Table)(({ theme }) => ({
  [`&.${tableClasses.root}`]: {
    marginLeft: theme.spacing(1),
    '& thead th': {
      paddingLeft: 10,
      fontSize: 18,
      fontWeight: '600',
      backgroundColor: '#17191a',
      color: theme.palette.common.white,
    },
    '& tbody td': {
      fontSize: 16,
      fontWeight: '300',
      backgroundColor: '#17191a',
      color: theme.palette.common.white,

    },
    '& tr:hover, td:hover': {
      backgroundColor: '#393d40',
    },
  },
}));

//styled paper
const StyledPaper = styled(Paper)(({ theme }) => ({
  [`&.${paperClasses.root}`]: {
    backgroundColor: '#17191a',
    color: theme.palette.common.white,
  },
}));

//styled popover
const StyledPopover = styled(Popover)(({ theme }) => ({
  [`&.${popoverClasses.root}`]: {
    // backgroundColor: '#17191a',
    color: "#17191a",
  },
}));

//syled topography with rounded corners
const StyledPopoverTypography = styled(Typography)(({ theme }) => ({
  [`&.${typographyClasses.root}`]: {
    backgroundColor: '#FFA836',
  },
}));

export default Containers
