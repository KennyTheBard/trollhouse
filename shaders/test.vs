#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

uniform vec4 anim[2];

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
            aux.x += anim[skin1].x * 0.5 + anim[skin2].x * 0.5;
            aux.y += anim[skin1].y * 0.5 + anim[skin2].y * 0.5;
            aux.z += anim[skin1].z * 0.5 + anim[skin2].z * 0.5;
            rot += anim[skin1][3] * 0.5 + anim[skin2][3] * 0.5;
            
        } else{
            if (skin1 >= 0) {
                aux.x += anim[skin1].x;
                aux.y += anim[skin1].y;
                aux.z += anim[skin1].z;
                rot += anim[skin1][3];
            } else {
                aux.x += anim[skin2].x;
                aux.y += anim[skin2].y;
                aux.z += anim[skin2].z;
                rot += anim[skin2][3];
            }
        }
    }
    gl_Position = projection * camera * model * rotationY(rot) * aux;
}