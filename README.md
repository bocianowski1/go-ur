# Go UR

## UR20 Sim

`docker run --rm -it -p 5900:5900 -p 6080:6080 -p 29999-30004:29999-30004 --platform linux/amd64 --privileged -e ROBOT_MODEL=UR20 --name ursim universalrobots/ursim_e-series "control_log"`
