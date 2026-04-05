import {useEffect, useRef, useState} from 'react';
import './App.css';
import {CreateInfrastructure, DestroyInfrastructure, IsInfrastructureDeployed} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime/runtime";
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, LinearProgress, Paper, Stack, Switch, Tooltip, Typography} from "@mui/material";
import {styled} from "@mui/material/styles";
import {Visibility, VisibilityOff} from "@mui/icons-material";

const OrchestrationSwitch = styled(Switch)(() => ({
    width: 128 * 3,
    height: 62 * 3,
    padding: 6 * 3,
    "& .MuiSwitch-switchBase": {
        padding: 6 * 3,
        transitionDuration: "300ms",
        "&.Mui-checked": {
            transform: "translateX(calc(70px * 3))",
            color: "#0f2c3a",
            "& + .MuiSwitch-track": {
                backgroundColor: "#9fe6ff",
                opacity: 1,
            },
        },
        "&.Mui-disabled": {
            color: "#d4d4d4",
            "& + .MuiSwitch-track": {
                backgroundColor: "#d4d4d4",
                opacity: 0.9,
            },
        },
    },
    "& .MuiSwitch-thumb": {
        boxSizing: "border-box",
        width: 50 * 3,
        height: 50 * 3,
        backgroundColor: "#f7fbff",
        boxShadow: "0 8px 18px rgba(20, 60, 80, 0.25)",
        border: "1px solid rgba(15, 44, 58, 0.08)",
        position: "relative",
        "&::before": {
            content: "\"\"",
            position: "absolute",
            width: "58%",
            height: "58%",
            top: "21%",
            left: "21%",
            backgroundImage:
                "url(\"data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2024%2024'%20fill='none'%20stroke='%230f2c3a'%20stroke-width='2'%20stroke-linecap='round'%20stroke-linejoin='round'%3E%3Cpath%20d='M12%202v10'/%3E%3Cpath%20d='M7.5%204.5a8%208%200%201%200%209%200'/%3E%3C/svg%3E\")",
            backgroundRepeat: "no-repeat",
            backgroundPosition: "center",
            backgroundSize: "contain",
        },
    },
    "& .MuiSwitch-track": {
        borderRadius: 999,
        backgroundColor: "#d7effa",
        opacity: 1,
        boxShadow: "inset 0 0 0 1px rgba(15, 44, 58, 0.08)",
    },
    "& .MuiSwitch-switchBase.Mui-disabled + .MuiSwitch-track": {
        boxShadow: "inset 0 0 0 1px rgba(15, 44, 58, 0.06)",
    },
}));

const OrchestrationProgress = styled(LinearProgress)(() => ({
    height: 10,
    borderRadius: 999,
    backgroundColor: "#d7effa",
    boxShadow: "inset 0 0 0 1px rgba(15, 44, 58, 0.08)",
    "& .MuiLinearProgress-bar": {
        borderRadius: 999,
        background: "linear-gradient(90deg, #9fe6ff 0%, #5ad1ff 100%)",
        boxShadow: "0 6px 12px rgba(20, 60, 80, 0.18)",
    },
}));

function App() {
    const [resultText, setResultText] = useState("");
    const [logLines, setLogLines] = useState<string[]>([]);
    const [isEnabled, setIsEnabled] = useState(false);
    const [isBusy, setIsBusy] = useState(false);
    const [isInitializing, setIsInitializing] = useState(true);
    const [progressValue, setProgressValue] = useState(0);
    const [displayProgress, setDisplayProgress] = useState(0);
    const [showLogs, setShowLogs] = useState(false);
    const [ssoDeviceCode, setSsoDeviceCode] = useState<string | null>(null);
    const [pendingCreate, setPendingCreate] = useState(false);
    const [pendingDestroy, setPendingDestroy] = useState(false);
    const displayProgressRef = useRef(0);
    const logPanelRef = useRef<HTMLDivElement | null>(null);
    const isEnabledRef = useRef(isEnabled);
    const isBusyRef = useRef(isBusy);
    const pendingCreateRef = useRef(pendingCreate);
    const pendingDestroyRef = useRef(pendingDestroy);

    useEffect(() => {
        displayProgressRef.current = displayProgress;
    }, [displayProgress]);

    useEffect(() => {
        isEnabledRef.current = isEnabled;
    }, [isEnabled]);

    useEffect(() => {
        isBusyRef.current = isBusy;
    }, [isBusy]);

    useEffect(() => {
        pendingCreateRef.current = pendingCreate;
    }, [pendingCreate]);

    useEffect(() => {
        pendingDestroyRef.current = pendingDestroy;
    }, [pendingDestroy]);

    useEffect(() => {
        if (!showLogs) {
            return;
        }
        const frame = requestAnimationFrame(() => {
            const panel = logPanelRef.current;
            if (panel) {
                panel.scrollTop = panel.scrollHeight;
            }
        });
        return () => cancelAnimationFrame(frame);
    }, [logLines, showLogs]);

    useEffect(() => {
        let animationFrame = 0;
        const startValue = displayProgressRef.current;
        const targetValue = progressValue;
        const delta = Math.abs(targetValue - startValue);
        const duration = Math.max(200, Math.min(800, delta * 8));
        let startTime: number | null = null;

        const step = (timestamp: number) => {
            if (startTime === null) {
                startTime = timestamp;
            }

            const elapsed = timestamp - startTime;
            const t = Math.min(elapsed / duration, 1);
            const nextValue = startValue + (targetValue - startValue) * t;
            displayProgressRef.current = nextValue;
            setDisplayProgress(nextValue);

            if (t < 1) {
                animationFrame = requestAnimationFrame(step);
            }
        };

        animationFrame = requestAnimationFrame(step);

        return () => cancelAnimationFrame(animationFrame);
    }, [progressValue]);

    useEffect(() => {
        const cancelLog = EventsOn("orchestrationLog", (message: string, progress?: number) => {
            setLogLines((prev) => [...prev, message]);
            if (typeof progress === "number") {
                setProgressValue(Math.max(0, Math.min(100, progress)));
            }

            if (
                pendingCreateRef.current &&
                (message.includes("Browser authentication successful") ||
                    message.includes("Successfully obtained AWS credentials from SSO"))
            ) {
                setResultText("Creating infrastructure, please wait...");
            }

            if (
                pendingDestroyRef.current &&
                (message.includes("Browser authentication successful") ||
                    message.includes("Successfully obtained AWS credentials from SSO"))
            ) {
                setResultText("Destroying infrastructure, please wait...");
            }
        });

        const cancelSso = EventsOn("ssoDeviceCode", (code: string) => {
            setSsoDeviceCode(code);
        });

        const initialize = async () => {
            setIsInitializing(true);
            setProgressValue(0);
            setResultText("Setting up, this may take a minute...");

            try {
                const deployed = await IsInfrastructureDeployed();
                setIsEnabled(deployed);
                setResultText("");
            } catch (error) {
                setResultText("Initialization failed. Check logs for details.");
            } finally {
                setIsInitializing(false);
            }
        };

        void initialize();

        return () => {
            cancelLog();
            cancelSso();
        };
    }, []);

    async function toggleInfrastructure(nextEnabled: boolean) {
        if (isBusy || isInitializing) {
            return;
        }

        const previous = isEnabled;
        setIsEnabled(nextEnabled);
        setIsBusy(true);
        setProgressValue(0);

        try {
            if (nextEnabled) {
                setPendingCreate(true);
                setResultText("Waiting for AWS SSO login...");
                const result = await CreateInfrastructure();
                setResultText(`Infrastructure created! Deployment URL: ${result}`);
            } else {
                setPendingDestroy(true);
                setResultText("Waiting for AWS SSO login...");
                await DestroyInfrastructure();
                setResultText("Infrastructure destroyed!");
            }
        } catch (error) {
            setIsEnabled(previous);
            setResultText("Operation failed. Check logs for details.");
        } finally {
            setPendingCreate(false);
            setPendingDestroy(false);
            setIsBusy(false);
        }
    }

    return (
        <div id="App">
            <Paper className="orchestration-card" elevation={0}>
                <Stack spacing={2.5} alignItems="center">
                    <Typography variant="h3" className="headline">
                        DeVILStarter
                    </Typography>
                    <Typography variant="body2" className="subhead">
                        {isInitializing
                            ? "Initializing..."
                            : isEnabled
                                ? "Click to destroy DeVILSona's infrastructure!"
                                : "Click to provision DeVILSona's infrastructure!"}
                    </Typography>
                    <Box className="toggle-shell">
                        <OrchestrationSwitch
                            checked={isEnabled}
                            disabled={isBusy || isInitializing}
                            onChange={(event) => toggleInfrastructure(event.target.checked)}
                        />
                    </Box>
                    <Typography className="status-text" variant="body2">
                        {resultText}
                    </Typography>
                    {(isInitializing || isBusy) && (
                        <Box sx={{width: "100%"}}>
                            <Box sx={{display: "flex", alignItems: "center", gap: 1}}>
                                <Box sx={{flexGrow: 1}}>
                                    <OrchestrationProgress variant="determinate" value={displayProgress} />
                                </Box>
                                <Typography variant="caption" sx={{color: "#0f2c3a", minWidth: 42, textAlign: "right"}}>
                                    {Math.round(displayProgress)}%
                                </Typography>
                            </Box>
                        </Box>
                    )}
                    <Box sx={{width: "100%", display: "flex", alignItems: "center", justifyContent: "space-between"}}>
                        <div className="connection-indicator">
                            <span
                                className={`connection-dot ${isInitializing || isBusy ? "connection-pending" : isEnabled ? "connection-on" : "connection-off"}`}
                            />
                            <span className="connection-text">
                                {isInitializing
                                    ? "Getting infra status"
                                    : isBusy
                                        ? isEnabled
                                            ? "Deploying infrastructure"
                                            : "Destroying infrastructure"
                                        : isEnabled
                                            ? "Infrastructure deployed"
                                            : "Infrastructure destroyed"}
                            </span>
                        </div>
                        <Tooltip title={showLogs ? "Hide logs" : "Show logs"}>
                            <Button
                                aria-label={showLogs ? "Hide logs" : "Show logs"}
                                onClick={() => setShowLogs((prev) => !prev)}
                                variant="outlined"
                                startIcon={showLogs ? <VisibilityOff /> : <Visibility />}
                                sx={{
                                    borderRadius: 999,
                                    textTransform: "none",
                                    borderColor: "rgba(15, 44, 58, 0.2)",
                                    color: "rgba(15, 44, 58, 0.85)",
                                    px: 2,
                                    py: 0.5,
                                    "&:hover": {
                                        borderColor: "rgba(15, 44, 58, 0.35)",
                                        backgroundColor: "rgba(15, 44, 58, 0.04)",
                                    },
                                }}
                            >
                                {showLogs ? "Hide logs" : "Show logs"}
                            </Button>
                        </Tooltip>
                    </Box>
                </Stack>
            </Paper>
            <Dialog
                open={showLogs}
                onClose={() => setShowLogs(false)}
                fullWidth
                maxWidth="md"
            >
                <DialogTitle>Orchestration Logs</DialogTitle>
                <DialogContent
                    dividers
                    sx={{height: "60vh", p: 2, pb: 1, display: "flex", overflow: "hidden"}}
                >
                    <div
                        id="orchestration-log"
                        className="log-panel"
                        ref={logPanelRef}
                        style={{height: "100%", width: "100%"}}
                    >
                        {logLines.length === 0 ? "Orchestration logs will appear here." : logLines.join("\n")}
                    </div>
                </DialogContent>
                <DialogActions sx={{pt: 1, px: 2}}>
                    <Button onClick={() => setShowLogs(false)} variant="outlined">
                        Close
                    </Button>
                </DialogActions>
            </Dialog>
            <Dialog
                open={Boolean(ssoDeviceCode)}
                onClose={() => setSsoDeviceCode(null)}
                fullWidth
                maxWidth="xs"
            >
                <DialogTitle>AWS SSO Verification Code</DialogTitle>
                <DialogContent dividers>
                    <Typography variant="body2" sx={{color: "rgba(15, 44, 58, 0.7)", mb: 1}}>
                        Use this code to complete AWS SSO authentication.
                    </Typography>
                    <Box
                        sx={{
                            width: "100%",
                            textAlign: "center",
                            fontSize: 24,
                            fontWeight: 700,
                            letterSpacing: "0.2em",
                            color: "#0f2c3a",
                            backgroundColor: "rgba(15, 44, 58, 0.06)",
                            border: "1px solid rgba(15, 44, 58, 0.1)",
                            borderRadius: 2,
                            py: 2,
                        }}
                    >
                        {ssoDeviceCode}
                    </Box>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setSsoDeviceCode(null)} variant="outlined">
                        Close
                    </Button>
                </DialogActions>
            </Dialog>
        </div>
    )
}

export default App
