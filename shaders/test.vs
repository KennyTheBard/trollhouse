#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

uniform vec3 anim[2];

in vec3 vert;
in vec2 vertTexCoord;
in float skinAttr;

out vec2 fragTexCoord;

void main() {
    fragTexCoord = vertTexCoord;
    int skin = int(skinAttr);

    vec4 aux = vec4(vert, 1);
    if (skin >= 0) {
        aux.x += anim[skin].x;
        aux.y += anim[skin].y;
        aux.z += anim[skin].z;
    }
    gl_Position = projection * camera * model * aux;
}