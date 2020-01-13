#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

uniform vec3 animT[2];
uniform float animR[2];
uniform vec3 animS[2];

in vec3 vert;
in vec2 vertTexCoord;
in vec2 skinAttr;

out vec2 fragTexCoord;

mat4 rotationY(float rad) {
	return mat4(	cos(rad),		0,		sin(rad),	0,
			 				0,		1.0,			 0,	0,
					-sin(rad),	0,		cos(rad),	0,
							0, 		0,				0,	1);
}

void main() {
    fragTexCoord = vertTexCoord;
    int skin1 = int(skinAttr[0]);
    int skin2 = int(skinAttr[1]);

    vec4 aux = vec4(vert, 1);
    float rot = 0.0;
    if (skin1 >= 0 || skin2 >= 0) {
        if (skin1 >= 0 && skin2 >= 0) {
            aux.x += animT[skin1].x * 0.5 + animT[skin2].x * 0.5;
            aux.y += animT[skin1].y * 0.5 + animT[skin2].y * 0.5;
            aux.z += animT[skin1].z * 0.5 + animT[skin2].z * 0.5;
            rot += animR[skin1] * 0.5 + animR[skin2] * 0.5;
            aux.x *= animS[skin1].x * 0.5 + animS[skin2].x * 0.5;
            aux.y *= animS[skin1].y * 0.5 + animS[skin2].y * 0.5;
            aux.z *= animS[skin1].z * 0.5 + animS[skin2].z * 0.5;
            
        } else{
            if (skin1 >= 0) {
                aux.x += animT[skin1].x;
                aux.y += animT[skin1].y;
                aux.z += animT[skin1].z;
                rot += animR[skin1];
                aux.x *= animS[skin1].x;
                aux.y *= animS[skin1].y;
                aux.z *= animS[skin1].z;
            } else {
                aux.x += animT[skin2].x;
                aux.y += animT[skin2].y;
                aux.z += animT[skin2].z;
                rot += animR[skin2];
                aux.x *= animS[skin2].x;
                aux.y *= animS[skin2].y;
                aux.z *= animS[skin2].z;
            }
        }
    }
    gl_Position = projection * camera * model * rotationY(rot) * aux;
}